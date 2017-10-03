# captionbot

Captionbot is a simple API wrapper for https://www.captionbot.ai/

## Installation

`go get github.com/nhatbui/captionbot`

## Usage

```
package main

import (
	"fmt"
	"github.com/nhatbui/captionbot"
)

func main() {
	bot := captionbot.CaptionBot{}
	// DON'T FORGET TO INITIALIZE!!!
	bot.Initialize()

	imgURL := "http://www.nhatqbui.com/assets/me.jpg"

	fmt.Println(bot.URLCaption(imgURL))

	// Or upload it

	imgFile := "/path/to/image.jpg"
	fmt.Println(bot.UploadCaption(imgFile))
}
```

## Thanks

Thanks to @krikunts for their work on [captionbot in Python](https://github.com/krikunts/captionbot) that inspired this package.
