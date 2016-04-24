#! /usr/bin/perl
 use Time::HiRes;
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

my $c=0;
my $max=($params>0)? $param: 10;
while (<STDIN>) {
    my $signal = $_;
    chomp $signal;
    #Using $signal as a number to sleep for testing purposes.
    Time::HiRes::usleep($signal);
    print price($lang,$params,$signal),"\n";
    break if ($c==$max);
    $c++;
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
