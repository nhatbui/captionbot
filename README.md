# captionbot

Captionbot is a simple API wrapper for https://www.captionbot.ai/

## Installation

`go get github.com/nhatbui/captionbot`

## Usage

```
package main

import (
	"fmt"
	"os"

	"github.com/nhatbui/captionbot"
)

func main() {
	bot, err := captionbot.New()
	if err != nil {
		fmt.Errorf("error instantiating bot %s", err)
		os.Exit(-1)
	}

	imgURL := "http://www.nhatqbui.com/assets/me.jpg"

	caption, err := bot.URLCaption(imgURL)
	if err != nil {
		fmt.Errorf("error uploading caption %s", err)
		os.Exit(-1)
	}
	fmt.Println(caption)

	// Or upload it

	imgFile := "/path/to/image.jpg"
	caption, err = bot.UploadCaption(imgFile)
	if err != nil {
		fmt.Errorf("error uploading caption %s", err)
		os.Exit(-1)
	}
	fmt.Println(caption)
}
```

## Thanks

Thanks to @krikunts for their work on [captionbot in Python](https://github.com/krikunts/captionbot) that inspired this package.
