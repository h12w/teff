Test Friendly Format (TEFF)
===========================

Copyright © 2014-2015 Hǎiliàng Wáng.

This specification is licensed under the [Creative Commons Attribution 4.0
International License](http://creativecommons.org/licenses/by/4.0/).

Introduction
------------
TEFF (TEst Friendly Format) is an extensible data format with testing purpose in
mind. It is friendly to read, write and compare, and can be extended to represent
rich set of data types.

In general, TEFF are organized into two layers, core and extensions. The
[core](#core) of TEFF represents a tree with annotated nodes, which forms
an extensible foundation with minimal constraints. The [extensions](#extensions)
of TEFF define encodings of major data types and allow custom encoding of user
defined types. The encodings of extensions are only constrained by the core, i.e.
any two extensions may have the same representation without causing any conflict.

This specification is a follow-up work of [OGDL 2.0](https://github.com/ogdl)
([OGDL](http://ogdl.org/) was invented by Rolf Veen, and we cooperated in
writing its 2.0 spec).

Notation
--------
The syntax is specified using a variant of Extended Backus-Naur Form, based on
[W3C XML EBNF](http://www.w3.org/TR/xml11/#sec-notation),
which is extended with the following definitions:
* Escape sequences defined in section [Escape sequences](#escape-sequences).
* Regular expressions defined in section [Regular expression](#regular-expression).
* Text enclosed by <> is a description.

Core
----

### Characters

A TEFF file is a sequence of [Unicode](http://unicode.org/) code points encoded
in [UTF-8](http://www.unicode.org/versions/latest/ch03.pdf).

Only `char_valid`, i.e. `\t` (U+0009), `\n` (U+000A), `\r` (U+000D) and code
points larger or equal to U+0020 (space) are considered valid in a TEFF file.

    char_valid     ::= char_inline | char_break
    char_inline    ::= char_visible | char_space
    char_visible   ::= [^\x00-\x20]
    char_space     ::= [ \t]
    char_break     ::= [\r\n]

### Lines

A TEFF file is also a sequence of lines separated by `newline`.

    line           ::= empty_line | content_line
    empty_line     ::= char_space* newline
    newline        ::= char_break | "\r\n" | EOF
    EOF            ::= <end of file>
    content_line   ::= indent_space (annotation | reference | value) newline
    indent_space   ::= char_space*
    annotation     ::= "#" char_inline*
    reference      ::= "^" char_inline*
    value          ::= [^\x00-\x20#^] char_inline*

### Indents

    start          ::= indent
    end            ::= unindent

Tokens `indent` and `unindent` are emitted by the rules described below:

1. A stack is used to store `indent_space` and controls the emission of
   `indent` & `unindent` tokens.
2. Initially, an empty value is pushed onto the stack, and then the TEFF file is
   scanned line by line to get the `indent_space` of each line.
3. When the top of the stack is the same as the `indent_space` of the current
   line, neither `indent` nor `unindent` is emitted.
4. When the top of the stack is a prefix of the `indent_space` of the current
   line, the `indent_space` is pushed onto the stack, and an `indent` token is
   emitted.
5. When the `indent_space` of the current line is the same as one of the non-top
   elements of the stack, the top of the stack is popped and an `unindent` token
   is emitted until the non-top element becomes the top. The number of `unindent`
   tokens emitted is the same as the number of elements popped.
6. If none of 3 to 5 happens, a syntax error occurs.
7. When `EOF` is emitted but the length of the stack is larger than 1, the top of
   the stack is popped and an `unindent` token is emitted until the length of
   the stack becomes 1.

### Grammar

    teff_file    ::= list EOF
    list         ::= node*
    node         ::= annotation* (value_list | reference)
    value_list   ::= value (start list end)?

Extensions
----------
In this section, extensions for `annotation`, `reference`, `list` and `value`
are defined to represent major data types, including almost all built-in types
and some of the important types in the standard libraries.

### Type annotation

TEFF can optionally specify data types by type annotations.

    type_annotation ::= "#" spaces? type_label
    ---------------     --- ------------------
        ↓                ↓       ↓
    ----------          --- ------------
    annotation      ::= "#" char_inline*

    type_label      ::= "<" letter_digit+ ">"
    unicode_letter  ::= <a Unicode code point classified as "Letter">
    unicode_digit   ::= <a Unicode code point classified as "Decimal Digit">
    letter_digit    ::= unicode_letter | unicode_digit | "_"

### Reference

TEFF can represent a cyclic graph by references. A reference is an absolute path
from the root node to one of its descendants.

The reference of the root object is `^` itself.

Each level of path is represented with `ref_segment` that depends on the type of
the parent object of the `seg_segment`.

    reference      ::= "^" (ref_segment)*
    ---------          --- --------------
        ↓               ↓       ↓
    ---------          --- ------------
    reference      ::= "^" char_inline*

And the specific definition of `ref_segment` depends on the parent type, e.g.
`array` or `map`.

### Array

An array is represented as a list.

    array         ::= array_element*
    -----             --------------
     ↓                  ↓
    ----              -----
    list          ::= node*

To represent an array of array, the anonymous symbol `_` is introduced to
represent the anonymous parent of a child array.

    array_element ::= "_"    start array end
    -------------     ---    ----- ----- ---
     ↓                 ↓       ↓    ↓     ↓
    ----------        -----  ----- ----  ---
    value_list    ::= value (start list  end)?

e.g.

    -
        1
        2
        3
    -
        4
        5

The `ref_segment` for a child of an array is defined as below:

    ref_segment   ::= "[" array_index "]"
    array_index   ::= decimals

### Map
A map is represented with a list of key-value pairs. Each pair is represented as
a node.

    map        ::= key_value*
    ---            ----------
     ↓               ↓
    ----           -----
    list       ::= node*

The key in a key-value pair is encoded a `value` suffixed by a `:`, and the
value in a key-value pair is encoded as a `list`.

    key_value  ::= map_key ":" start map_value end
    ---------      ----------- ----- --------- ---
     ↓               ↓           ↓      ↓       ↓
    ----------     -----       -----   ----    ---
    value_list ::= value      (start   list    end)?

Encoding of `map_key`:

* identifier: [`raw_string`](#string)
* string: [`interpreted_string`](#string)
* boolean: [`boolean`](#boolean-value)
* numeric: [`numeric`](#numeric-value)
* others: implementation specific, as long as the encoding satisfies `value` and
the ending of the encoding is recognized without relying on the `:`.

The `ref_segment` for a child of a map depends on its `map_key`.

When the `map_key` is an identifier:

    ref_segment   ::= "[" array_index "]"
    array_index   ::= decimals



### Nil

The special `value` nil is used to represent an uninitialized nullable node.

    nil   ::= "nil"
    ---       -----
     ↓          ↓
    -----     ---------------------------
    value ::= [^\x00-\x20#^] char_inline*

### String
A string is represented as either a `raw_string` or an `interpreted_string` (double
quoted).

    string             ::= raw_string | interpreted_string
    ------                 -------------------------------
      ↓                               ↓
    -----                  ---------------------------
    value              ::= [^\x00-\x20#^] char_inline*

A string value can be represented as a `raw_string` if and only if:

* It is not empty.
* It does not starts with `char_space`, `#` or `^`.
* It only contains `char_inline`.

    raw_string         ::= value

An `interpreted_string` is quoted with double quotes `"` and can contain any
bytes by escape sequences.

    quoted_char        ::= (char_inline - '"') | '\\"'

    interpreted_string ::= '"' quoted_char* '"'
    ------------------     --------------------
      ↓                             ↓
    -----                  ---------------------------
    value              ::= [^\x00-\x20#^] char_inline*

#### Escape sequences

    \a    U+0007 alert or bell
    \b    U+0008 backspace
    \t    U+0009 horizontal tab
    \n    U+000A line feed or newline
    \v    U+000B vertical tab
    \f    U+000C form feed
    \r    U+000D carriage return
    \\    U+005C backslash
    \"    U+0022 double quote "
    \x    Any byte represented with two hexadecimal digits followed by \x
    \u    Unicode code point represented with exactly 4 hexadecimal digits followed by \u
    \U    Unicode code point represented with exactly 8 hexadecimal digits followed by \U

### Regular expression
A regular expression is a `value`. The syntax of regular expressions are
defined by [Golang Regexp](http://golang.org/pkg/regexp/syntax/).

### Boolean value
Boolean value is a `value` of either true of false.

    boolean ::= "true" | "false"
    -------     ----------------
      ↓                ↓
    -----       ---------------------------
    value   ::= [^\x00-\x20#^] char_inline*

### Numeric value
Numeric value is a `value` that encode a number.

    sign       ::= "+" | "-"
    decimals   ::= "0" | [1-9] [0-9]*

#### Integer
    integer    ::= sign? decimals
    -------        --------------
       ↓                 ↓
    -----          ---------------------------
    value      ::= [^\x00-\x20#^] char_inline*

#### Float
Float value is a `value` that encode a floating point number:

    exponent   ::= ( "e" | "E" ) ( "+" | "-" )? decimals
    float_base ::= (decimals "." decimal* exponent?) |
                   (decimals exponent) |
                   ("." decimals exponent?)

    float      ::= sign? float_base
    -----          ----------------
      ↓                   ↓
    -----          ---------------------------
    value      ::= [^\x00-\x20#^] char_inline*

#### Complex
    int_float  ::= decimals | float_base

    complex    ::= sign? int_float sign int_float "i"
    -------        ----------------------------------
      ↓                        ↓
    -----          ---------------------------
    value      ::= [^\x00-\x20#^] char_inline*

### Date/time (TODO: use a shorter representation)
A date/time value is an `value` encoded with
[RFC3339](http://www.rfc-editor.org/rfc/rfc3339.txt)

    date_time ::= rfc3339_date_time
    ---------     -----------------
      ↓                   ↓
    -----         ---------------------------
    value     ::= [^\x00-\x20#^] char_inline*

e.g.

    2006-01-02T15:04:05.999999999Z07:00

### IP address
An IP address is either an IPv4 or IPv6 address.

    ip    ::= ipv4 | ipv6

An IPv4 address value is an `value` encoded with dot-decimal notation:

    ipv4  ::= decimals "." decimals "." decimals "." decimals
    ----      -----------------------------------------------
     ↓                    ↓
    -----     ---------------------------
    value ::= [^\x00-\x20#^] char_inline*

e.g.

    74.125.19.99

An IPv6 address value is an `value` encoded with
[RFC5952](http://www.rfc-editor.org/rfc/rfc5952.txt).

    ipv6  ::= rfc5952_ipv6_address
    ----      --------------------
     ↓                 ↓
    -----     ---------------------------
    value ::= [^\x00-\x20#^] char_inline*

e.g.

    2001:4860:0:2001::68

### Multi-line String (TODO)

### Multi-line Regular Expressions (TODO)

### URL

### Custom extensions (TODO)
Custom encoding can be implemented as long as it does not conflict with the
built-in encodings.
