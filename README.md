# Modular development branch

This branch is trying to make the client modular by rewriting all the networking in go, and using the PureBasic application purely as frontend.
This has the following advantages:

- It will work headless.
- It can have support for captcha solving or tracking.
- It can retrieve things like fingerprints automatically.
- It can have built in proxy and VPN support. (That's a thing for later)
- It can support several similar websites/games. (Depending if someone writes a module for those games)
- It will be possible to run multiple instances at once.
- The PureBasic application will now serve as a central controller of all instances.
- Important parts can be compiled by anyone with a freely available compiler.

All the 'can' things depend on if and how they are implemented.
I'm not sure yet if the main logic (Checking, queueing and placing pixels) will be inside each 'client' instance, or if all clients depend on a central application somewhere.

# Pixelcanvas.io Custom Client

This is a custom client / bot for Pixelcanvas.io.
In contrast to similar projects this client is made to be configurable and usable without any hassles.
Programming experience isn't necessary.
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

## Usage

1. [Download](https://github.com/Dadido3/Pixelcanvas.io-Custom-Client/releases)
2. Start the client.
3. Click "Fingerprint" and enter your pixelcanvas.io fingerprint.
4. Add your templates into the "Templates" view. Activate templates by checking their checkbox. (If you haven't done that already)
5. Let the client do the work until a message requester with the title "Captcha" pops up. You then have to use the pixelcanvas.io website and place a pixel which causes a recaptcha to appear. Solve the captcha(s), click "OK" on the requester, and the bot will work for another hour.

## Known problems (Which also may never be fixed)

- If you zoom out and press "Load Viewport" the client may hit the process handle limit, as this will create a lot of small images.
  This could be fixed by caching or using bigger image chunks.
  (But meh, just don't let it load 10000 images ;) )
- Most network communication is blocking (Except chunk downloading). But that shouldn't cause any troubles.
- Estimations and pixel rates are wrong after starting the client, but they will show correct values after 30 placed pixels.
- There is probably more small stuff, but the client does its job.

## Screenshots

A client running for a month:
![<Image missing>](/screens/V0.946.png)
