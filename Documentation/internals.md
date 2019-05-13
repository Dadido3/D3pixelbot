# D3pixelbot internal architecture

## Chunk download mechanism

The chunk download mechanism is basing on frequent requests from listeners.
Each request contains a rectangle that the canvas should try to keep in sync with the game.

Not queried chunks will get unloaded/deleted automatically after some time.

While a chunk is downloading, all pixel events will be queued.
After is has been downloaded, all events will be replayed.
This will make sure that the data will not get out of sync while chunk data is being downloaded.

```mermaid
sequenceDiagram
    participant listener1
    participant listener2
    participant canvas
    participant game connection
    participant chunk

    loop goroutine
        listener1 ->> canvas: queryRect(rect) every minute
        note right of listener1: Or more frequent<br>if rectangles change
    end
    canvas ->> chunk: creates chunk if necessary. getQueryState(true)
    chunk ->> canvas: result: download or keep existing chunk
    canvas --x game connection: request chunk download
    game connection ->> canvas: signalDownload(rect)
    canvas ->> +chunk: signalDownload()
    canvas ->> listener1: handleSignalDownload(rect)
    canvas ->> listener2: handleSignalDownload(rect)
    game connection -->> canvas: setPixel(pos, color)
    canvas ->> chunk: setPixel(pos, color)
    loop goroutine
        game connection->>game connection: download chunk data
    end
    game connection ->> canvas: setImage(img, false)
    canvas ->> chunk: setImage(img)
    chunk ->> -canvas: result: image with replayed events
    canvas ->> listener1: handleSetImage(img)
    canvas ->> listener2: handleSetImage(img)
```