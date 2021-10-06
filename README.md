Gopicto is a command line tool to generate a PDF containing a picture and an associated word.
Inspired from PictoSelector.

# Usage

To get help:

```
./gopicto --help
```

To generate a PDF:

```
./gopicto -c config.sample.yaml -o /tmp/test.pdf
```

# Configuration file

Configuration file is using YAML and is as follow (see sample for a concrete example):

```
# Options related to a page
page:
  cols: <number of columns>
  lines: <number of lines>
  orientation <landscape|portrait>
  page_margins:
    top: <document top margin>
    bottom: <document bottom margin>
    left: <document left margin>
    right: <document right margin>
  margins:
    top: <top margin for each cell>
    bottom: <bottom margin for each cell>
    left: <left margin for each cell>
    right: <right margin for each cell>
  paddings:
    top: <top padding for each cell>
    bottom: <bottom padding for each cell>
    left: <left padding for each cell>
    right: <right padding for each cell>

# Options regarding text printed in the PDF
text:
  font: <path to a font> # font to use for the text, if not provided, use a default font
  ratio: <ratio> # The text part will take 'ratio' of an entire cell
  color: <color of the text>
  firstLetterColor: <color of the first letter of each cell's text>

# Options regarding images to put in the PDF
images:
  - image: <path to a local image>
    text: <text to display below the image>
  ...
```

