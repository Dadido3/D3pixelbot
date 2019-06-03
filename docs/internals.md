# D3pixelbot internal architecture

## Chunk download mechanism

Each listener can register an unlimited amount of rectangles it wants to listen to.
The canvas periodically queries the chunks based on the rectangles.
Based on the result of each query something of the following will happen:

- If the chunk is invalid, a download request will be sent to the game connection
- If the chunk hasn't been queried in a while, it will be deleted (TODO: or compressed)

While a chunk is downloading, all pixel events will be queued.
After the chunk has been downloaded, all events will be replayed.
This will make sure that the data will not get out of sync while chunk data is being downloaded.

```mermaid
sequenceDiagram
    participant listener1
    participant listener2
    participant canvas
    participant game connection
    participant chunk

    
    listener1 ->> canvas: registerRects(rect)
    loop canvas goroutine
        canvas ->> chunk: creates chunk if necessary. getQueryState(true)
        chunk ->> canvas: result: download or keep existing chunk
        canvas --x game connection: request chunk download
    end
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