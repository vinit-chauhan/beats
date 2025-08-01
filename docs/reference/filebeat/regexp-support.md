---
mapped_pages:
  - https://www.elastic.co/guide/en/beats/filebeat/current/regexp-support.html
applies_to:
  stack: ga
---

# Regular expression support [regexp-support]

Filebeat regular expression support is based on [RE2](https://godoc.org/regexp/syntax).

Filebeat has several configuration options that accept regular expressions. For example, `multiline.pattern`, `include_lines`, `exclude_lines`, and `exclude_files` all accept regular expressions. Some options, however, such as the input `paths` option, accept only glob-based paths.

Before using a regular expression in the config file, refer to the documentation to verify that the option you are setting accepts a regular expression.

::::{note}
We recommend that you wrap regular expressions in single quotation marks to work around YAML’s string escaping rules. For example, `'^\[?[0-9][0-9]:?[0-9][0-9]|^[[:graph:]]+'`.
::::


For more examples of supported regexp patterns, see [Managing Multiline Messages](/reference/filebeat/multiline-examples.md). Although the examples pertain to Filebeat, the regexp patterns are applicable to other use cases.

The following patterns are supported:

* [Single Characters](#single-characters)
* [Composites](#composites)
* [Repetitions](#repetitions)
* [Groupings](#grouping)
* [Empty Strings](#empty-strings)
* [Escape Sequences](#escape-sequences)
* [ASCII Character Classes](#ascii-character-classes)
* [Perl Character Classes](#perl-character-classes)

| Pattern | Description |
| --- | --- |
| $$$single-characters$$$**Single Characters** |  |
| `x` | single character |
| `.` | any character |
| `[xyz]` | character class |
| `[^xyz]` | negated character class |
| `[[:alpha:]]` | ASCII character class |
| `[[:^alpha:]]` | negated ASCII character class |
| `\d` | Perl character class |
| `\D` | negated Perl character class |
| `\pN` | Unicode character class (one-letter name) |
| `\p{{Greek}}` | Unicode character class |
| `\PN` | negated Unicode character class (one-letter name) |
| `\P{{Greek}}` | negated Unicode character class |
| $$$composites$$$**Composites** |  |
| `xy` | `x` followed by `y` |
| `x&#124;y` | `x` or `y` (prefer `x`) |
| $$$repetitions$$$**Repetitions** |  |
| `x*` | zero or more `x` |
| `x+` | one or more `x` |
| `x?` | zero or one `x` |
| `x{n,m}` | `n` or `n+1` or … or `m` `x`, prefer more |
| `x{n,}` | `n` or more `x`, prefer more |
| `x{{n}}` | exactly `n` `x` |
| `x*?` | zero or more `x`, prefer fewer |
| `x+?` | one or more `x`, prefer fewer |
| `x??` | zero or one `x`, prefer zero |
| `x{n,m}?` | `n` or `n+1` or … or `m` `x`, prefer fewer |
| `x{n,}?` | `n` or more `x`, prefer fewer |
| `x{{n}}?` | exactly `n` `x` |
| $$$grouping$$$**Grouping** |  |
| `(re)` | numbered capturing group (submatch) |
| `(?P<name>re)` | named & numbered capturing group (submatch) |
| `(?:re)` | non-capturing group |
| `(?i)abc` | set flags within current group, non-capturing |
| `(?i:re)` | set flags during re, non-capturing |
| `(?i)PaTTeRN` | case-insensitive (default false) |
| `(?m)multiline` | multi-line mode: `^` and `$` match begin/end line in addition to begin/end text (default false) |
| `(?s)pattern.` | let `.` match `\n` (default false) |
| `(?U)x*abc` | ungreedy: swap meaning of `x*` and `x*?`, `x+` and `x+?`, etc (default false) |
| $$$empty-strings$$$**Empty Strings** |  |
| `^` | at beginning of text or line (`m`=true) |
| `$` | at end of text (like `\z` not `\Z`) or line (`m`=true) |
| `\A` | at beginning of text |
| `\b` | at ASCII word boundary (`\w` on one side and `\W`, `\A`, or `\z` on the other) |
| `\B` | not at ASCII word boundary |
| `\z` | at end of text |
| $$$escape-sequences$$$**Escape Sequences** |  |
| `\a` | bell (same as `\007`) |
| `\f` | form feed (same as `\014`) |
| `\t` | horizontal tab (same as `\011`) |
| `\n` | newline (same as `\012`) |
| `\r` | carriage return (same as `\015`) |
| `\v` | vertical tab character (same as `\013`) |
| `\*` | literal `*`, for any punctuation character `*` |
| `\123` | octal character code (up to three digits) |
| `\x7F` | two-digit hex character code |
| `\x{{10FFFF}}` | hex character code |
| `\Q...\E` | literal text `...` even if `...` has punctuation |
| $$$ascii-character-classes$$$**ASCII Character Classes** |  |
| `[[:alnum:]]` | alphanumeric (same as `[0-9A-Za-z]`) |
| `[[:alpha:]]` | alphabetic (same as `[A-Za-z]`) |
| `[[:ascii:]]` | ASCII (same as `\x00-\x7F]`) |
| `[[:blank:]]` | blank (same as `[\t ]`) |
| `[[:cntrl:]]` | control (same as `[\x00-\x1F\x7F]`) |
| `[[:digit:]]` | digits (same as `[0-9]`) |
| `[[:graph:]]` | graphical (same as ```[!-~] == [A-Za-z0-9!"#$%&'()*+,\-./:;<=>?@[\\\]^_` ``` ```{&#124;}~]```) |
| `[[:lower:]]` | lower case (same as `[a-z]`) |
| `[[:print:]]` | printable (same as `[ -~] == [ [:graph:]]`) |
| `[[:punct:]]` | punctuation (same as ```[!-/:-@[-`{-~]```) |
| `[[:space:]]` | whitespace (same as `[\t\n\v\f\r ]`) |
| `[[:upper:]]` | upper case (same as `[A-Z]`) |
| `[[:word:]]` | word characters (same as `[0-9A-Za-z_]`) |
| `[[:xdigit:]]` | hex digit (same as `[0-9A-Fa-f]`) |
| $$$perl-character-classes$$$**Supported Perl Character Classes** |  |
| `\d` | digits (same as `[0-9]`) |
| `\D` | not digits (same as `[^0-9]`) |
| `\s` | whitespace (same as `[\t\n\f\r ]`) |
| `\S` | not whitespace (same as `[^\t\n\f\r ]`) |
| `\w` | word characters (same as `[0-9A-Za-z_]`) |
| `\W` | not word characters (same as `[^0-9A-Za-z_]`) |
