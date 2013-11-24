Configuration
==================
Make sure that there are no extra spaces or endlines in these files and that they are ANSI encoded.

Patterns.txt
-----------------------------
* Each line in <b>patterns.txt</b> corresponds to one match, that will be searched for.
* You have three options for one word when defining patterns:
  1. <b>TOKEN</b> (regular expression defined in tokens.txt) surrounded by <code><></code>
  2. <b>SPECIFIC WORD</b> surrounded by <code>{}</code>
  3. <b>ANYTHING</b> for that you can type _ and search for that will match anything
* Words on each line needs to be separated by spaces.
* Example line: <code>&lt;IP&gt; _ _ &lt;DATE&gt; {&quot;GET}</code>

Tokens.txt
-----------------------------
* One token definition per line like this: <code>NAME(single_space)regex</code>.
* The syntax of the regular expressions accepted is the same general syntax used by
Perl, Python, and other languages. 
More precisely, it is the syntax accepted by RE2 and described at http://code.google.com/p/re2/wiki/Syntax, except for \C.
