Pixelcanvas.io Custom Client
=====

This is a custom client / bot for Pixelcanvas.io.
Contrary to similar projects this client is made to be configurable and usable without any hassles.
Using this tool doesn't need any type of programming skills.
To create a job/template you just have to click "create" and change some settings (Image (PNG) path, position, placing strategy, ...) over the user interface.

As this project is just for the fun, some of the things that are considered possible future features may never be implemented.
These include:
- Captcha requester (Or even an interface to external captcha solving services)
- Managing of multiple users (Proxies, VPNs or possibly lightweight clients which connect to a "mother client")
- Manual pixel placement (But queued)
- Better queue reorder algorithms which are more human like
- Timelapse creator: It would be possible (And that was one of the other ideas i started that project with) to store every pixel change, and additionally store key frames of a big part of the canvas every few hours.
(These keyframes could easily be 10000x10000 pixels in size)
Afterwards the data could be used to create timelapses of nearly any part of the canvas at any given time interval.

## Known problems (Which also may never be fixed)
- If you zoom out and press "Load Viewport" the client may hit the process handle limit, as this will create a lot of small images.
  This could be fixed by caching or using bigger image chunks.
  (But meh, just don't let it load 10000 images ;) )
- Most network communication is blocking (Except chunk downloading). But that shouldn't cause any troubles.
- Estimations and pixelrates are wrong after starting the client, but they will show correct values after 30 placed pixels.
- There is probably more small stuff, but the client does its job.

## Screenshots
A client running for a month:
![<Image missing>](/Screenshots/V0.946.png)
