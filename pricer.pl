#! /usr/bin/perl
{
    my $ofh = select STDOUT;
	$| = 1;
	select $ofh;
}
{
    my $ofh = select STDIN;
	$| = 1;
	select $ofh;
}
my $lang = <STDIN>;
chomp $lang;

my $params = <STDIN>;
chomp $params;

while (<STDIN>) {
    my $signal = $_;
    chomp $signal;
    print price($lang,$params,$signal),"\n";
}

srand;
sub price {
    my $lang = shift;
    my $params = shift;
    my $signal = shift;

    return translate($lang)." : $params/$signal => ".rand(100);
}

sub translate {
    my $lang = shift;

    return "Hello" if ($lang eq 'EN');
    return "Bonjour" if ($lang eq 'FR');
}
