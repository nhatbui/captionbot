/*
Captionbot is a simple API wrapper for https://www.captionbot.ai/

Usage

        package main

        import (
                "fmt"
                "os"

                "github.com/nhatbui/captionbot"
        )

        func main() {

                bot, err := captionbot.New()
                if err != nil {
                        fmt.Printf("error instantiating bot %s\n", err)
                        os.Exit(1)
                }

                // caption a remote image by provideing a URL
                imgURL := "http://www.nhatqbui.com/assets/me.jpg"

                caption, err := bot.URLCaption(imgURL)
                if err != nil {
                        fmt.Printf("error uploading caption %s\n", err)
                        os.Exit(1)
                }
                fmt.Println(caption)

                // caption a local image by uploading it
                imgFile := "./sample.jpg"

                caption, err = bot.UploadCaption(imgFile)
                if err != nil {
                        fmt.Printf("error uploading caption %s\n", err)
                        os.Exit(1)
                }
                fmt.Println(caption)
        }
*/
package captionbot
