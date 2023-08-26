package html

import (
	"os"
	"path/filepath"
)

var Head = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <link rel="stylesheet" href="/style.css">
  </head>
  <body>
`

var Tail = `
  </body>
</html>
`

// AddCSS adds style.css file to dir.
func AddCSS(dir string) error {
	css := `
h2 { margin-bottom: 0px; }
h3 { margin-bottom: 0px; }
h4 { margin-bottom: 0px; }
h5 { margin-bottom: 0px; }

img 
{
max-width: 100%;
}

code
{
font-family: monospace;
}

pre
{
background: #f7f7f7;
border: 1px solid #d7d7d7;
margin: 1em 1.75em;
padding: .25em;
overflow: auto;
white-space: pre-wrap;
}

blockquote
{
font-family: cursive;
}

@media screen and (max-device-width: 480px)
{
	body
	{
	-webkit-text-size-adjust: none;
	}
}
`
	return os.WriteFile(filepath.Join(dir, "style.css"), []byte(css), 0640)
}
