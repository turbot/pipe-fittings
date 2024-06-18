package sperr

/*
Package sperr is used to store error chains.

sperr exposes an Error struct which is compatible with the std error interface

All error values returned from this package implement fmt.Formatter and can
be formatted by the fmt package. The following verbs are supported:

    %s    print the error. If the error has a Cause it will be
          printed recursively.
    %v    see %s
    %+v   extended format. Each Frame of the error's StackTrace will
          be printed in detail.

*/
