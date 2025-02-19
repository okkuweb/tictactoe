package tictactoe::Controller::Game;
use Mojo::Base 'Mojolicious::Controller', -signatures;
use Storable;

# This action will render a template
sub get_game {
    my $self = shift->openapi->valid_input or return;
    return $self->rendered(404) unless -f './games';

    my $games = retrieve('./games');
    my $key = $self->param('game_id');
    my %games = %{$games};
    if ($games{$key}) {
        $self->render(
            status => 200,
            json => $games{$key}
        );
    } else {
        $self->rendered(404);
    }
}

sub delete_game {
    my $self = shift->openapi->valid_input or return;
    return $self->rendered(404) unless -f './games';

    my $game_id = $self->param('game_id');
    my $games_ref = retrieve('./games');
    my %games = %{$games_ref};
    my $found;
    foreach my $key (keys %games) {
        if ($game_id eq $games{$key}->{id}) {
            $found = 1;
            delete $games{$key};
            store \%games, './games';
            last;
        }
    }
    if ($found) {
        $self->rendered(200);
    } else {
        $self->rendered(404);
    }
}

sub make_move {
    my $self = shift->openapi->valid_input or return;
    return $self->rendered(404) unless -f './games';

    my $games = retrieve('./games');
    my $hash = $self->req->json;
    my @player = split "", $hash->{board};

    my $game = $games->{$self->param('game_id')};
    my @server = split "", $game->{board};

    if (scalar @player != 9 || join("", @player) =~ /[^-XO]/ || $game->{status} ne "RUNNING") {
        $self->rendered(400);
        return;
    }

    my $check = 0;
    foreach (0 .. 8) {
        if ($player[$_] ne $server[$_]) {
            $check += 1;
            if ($player[$_] ne $game->{player} || ($player[$_] eq $game->{player} &&$player[$_] eq ($game->{player} eq "X" ? "O" : "X"))) {
                $self->rendered(400);
                return;
            }
        }
    }

    my $new_move;
    my $status = "RUNNING";
    if ($check > 1 || $check == 0) {
        $self->rendered(400)
    } else {
        if (my $new_state = $self->check_end(\@player)) {
            $status = $new_state;
        } else {
            ($new_move, $new_state) = $self->pc_move(\@player, $game->{player});
            if ($new_state) {
                $status = $new_state;
            }
        }
    }

    if ($new_move) {
        $games->{$self->param('game_id')}->{board} = join("", @{$new_move});
    } else {
        $games->{$self->param('game_id')}->{board} = join("", @player);
    }
    $games->{$self->param('game_id')}->{status} = $status;
    store $games, './games';
    $self->get_game;
}

sub check_end {
    my ($self, $move_ref) = @_;
    my @move = @{$move_ref};
    my $winner;

    my $current = $move[0];
    if ($current ne '-') {
        $winner = $current if $current eq $move[1] && $current eq $move[2];
        $winner = $current if $current eq $move[4] && $current eq $move[8];
        $winner = $current if $current eq $move[3] && $current eq $move[6];
    }
    $current = $move[8];
    if ($current ne '-') {
        $winner = $current if $current eq $move[7] && $current eq $move[6];
        $winner = $current if $current eq $move[5] && $current eq $move[2];
    }
    $current = $move[4];
    if ($current ne '-') {
        $winner = $current if $current eq $move[1] && $current eq $move[7];
        $winner = $current if $current eq $move[3] && $current eq $move[5];
    }

    if ($winner) {
        return $winner . "_WON";
    } else {
        return;
    }
}

sub pc_move {
    my ($self, $move_ref, $player_symbol) = @_;
    my @move = @{$move_ref};

    my $free = 0;
    foreach (@move) {
        if ($_ eq "-") {
            $free++;
        }
    }
    if ($free < 1) {
        return(\@move, "DRAW");
    }
    my $newmove = int rand $free;
    my $counter = 0;
    for (my $i = 0; $i < scalar @move; $i++) {
        if ($move[$i] eq "-") {
            if ($counter == $newmove) {
                $move[$i] = $player_symbol eq "X" ? "O" : "X"; 
            }
            $counter++;
        }
    }
    my $state = $self->check_end(\@move);
    return (\@move, $state);
}

1;

