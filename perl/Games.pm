package tictactoe::Controller::Games;
use Mojo::Base 'Mojolicious::Controller', -signatures;
use Data::GUID;
use Storable;

sub get_games {
    my $self = shift->openapi->valid_input or return;
    my @games;
    if (-f './games') {
        my $games = retrieve('./games');
        foreach my $key (keys %{$games}) {
            push @games, $games->{$key};
        }
    }
    $self->render(json => \@games, status => 200);
}

sub start_game {
    my $self = shift->openapi->valid_input or return;
    my $game_uuid = lc Data::GUID->new;
    my $symbol = int rand 2 ? "X" : "O";

    my $game;
    if (int rand 2) {
        my $start_move = int(rand(9));
        my @arr = qw/- - - - - - - - -/;
        $arr[$start_move] = $symbol eq "X" ? "O" : "X";
        $game = join "", @arr;
    } else {
        $game = "---------";
    }
    my %game = (
        id => $game_uuid,
        board => $game,
        status => "RUNNING",
        player => $symbol
    );

    my %games;
    if (-f "./games") {
        my $games_ref = retrieve('./games');
        %games = %{$games_ref};
    }
    $games{$game_uuid} = \%game;

    store \%games, './games';

    $self->res->headers->location($game_uuid);
    $self->rendered(201);
}

1;

