# Goldmark-Embed

This project is intended to provide a generic embed option for [Yuin/Goldmark](https://github.com/yuin/goldmark). While functional, this is not complete and can currently just inline embed video elements.

## Use

In golang:

```golang
import (
    "bytes"
    "github.com/yuin/goldmark"
    embed "github.com/zincarla/goldmark-embed"
)

md := goldmark.New(
          goldmark.WithExtensions(embed.DefaultEmbed),
      )
var buf bytes.Buffer
if err := md.Convert(source, &buf); err != nil {
    panic(err)
}
```

In markdown:

```markdown
?[]("mime/format" "./some/path.mp4")
```

Additional formats can be added to same statement like

```markdown
?[]("video/mp4" "./path.mp4" "video/ogg" "./path.ogg")
```

## Goldmark Extension dev notes

Things I learned about goldmark extensions. You have 3 main things you need to create. An AST node, a renderer, and a parser.

The parser is responsible for turning the markdown in an AST node. Your parser's Trigger function should return a slice of bytes that should Trigger your parser. This slice matches one byte for the inline parser. If you have multiple bytes in the slice, you will get triggers for each byte. You only need to match on the first byte of your node, and as part of the parser decide if the node is valid or not. When triggered, your parser's Parse function is called by goldmark. The reader that is passed to the function starts where the trigger ocurred in the markdown. When you parse a node, you should ensure the reader is seeked to the end of your node's markdown data before ending your function.

An AST node should be a representation of your markdown node/element. Whatever core data you need for your node, this object should contain it. A video for example, should include the mime and the url, or for an image, it should contain the caption and url. This AST node with your custom properties is passed to the renderer and there you will actually parse it into html. The dump function is for debugging purposes.

The renderer converts the data in your AST node, into actual html.
