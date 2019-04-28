# Pixelcanvas Reverse Engineering information

## General information

- Chunk size: 64x64 pixels
- Chunk collection: 15x15 chunks

Color Table:

``` PureBasic
Palette(00)\R = 255 : Palette(00)\G = 255 : Palette(00)\B = 255
Palette(01)\R = 228 : Palette(01)\G = 228 : Palette(01)\B = 228
Palette(02)\R = 136 : Palette(02)\G = 136 : Palette(02)\B = 136
Palette(03)\R = 034 : Palette(03)\G = 034 : Palette(03)\B = 034
Palette(04)\R = 255 : Palette(04)\G = 167 : Palette(04)\B = 209
Palette(05)\R = 229 : Palette(05)\G = 000 : Palette(05)\B = 000
Palette(06)\R = 229 : Palette(06)\G = 149 : Palette(06)\B = 000
Palette(07)\R = 160 : Palette(07)\G = 106 : Palette(07)\B = 066
Palette(08)\R = 229 : Palette(08)\G = 217 : Palette(08)\B = 000
Palette(09)\R = 148 : Palette(09)\G = 224 : Palette(09)\B = 068
Palette(10)\R = 002 : Palette(10)\G = 190 : Palette(10)\B = 001
Palette(11)\R = 000 : Palette(11)\G = 211 : Palette(11)\B = 221
Palette(12)\R = 000 : Palette(12)\G = 131 : Palette(12)\B = 199
Palette(13)\R = 000 : Palette(13)\G = 000 : Palette(13)\B = 234
Palette(14)\R = 207 : Palette(14)\G = 110 : Palette(14)\B = 228
Palette(15)\R = 130 : Palette(15)\G = 000 : Palette(15)\B = 128
```

## HTTPS Post methods

- Draw pixel:
  - URL: `https://pixelcanvas.io/api/pixel`

  - Body: JSON Object:

    | Key | Type | Description | Example |
    | --- | ---- | ----------- | ------- |
    | x             | int
    | y             | int
    | a             | int | `= x + y + 8`
    | color         | int | index of color
    | fingerprint   | string
    | token         | null

  - Result:JSON Object:

    | Key | Type | Description | Example |
    | --- | ---- | ----------- | ------- |
    | success       | bool
    | waitSeconds   | float | Seconds until next pixel can be drawn
    | errors        | array | contains `{msg: ...}` objects

    - Possible error msg strings:
      - `You are using a proxy!!!11!one`
      - `You must provide a token`
      - `You must wait`

- Me authentication
  - URL: `https://pixelcanvas.io/api/me`

  - Body: JSON Object:

    | Key | Type | Description | Example |
    | --- | ---- | ----------- | ------- |
    | fingerprint   | string
  
  - Result:JSON Object:

    | Key | Type | Description | Example |
    | --- | ---- | ----------- | ------- |
    | id            | string | | `ip:12.34.56.78`
    | name          | string | | `Anonymous`
    | center        | array | x, y coordinates of your center | `[0,2000]`
    | waitSeconds   | float | Seconds until next pixel can be drawn

## HTTPS Get methods

- Download chunk collection (15x15 chunks):

  - URL: `https://pixelcanvas.io/api/bigchunk/<ccx>.<ccy>.bmp`  
    With `ccx` and `ccy` being the offset of the chunk collection.
    Example: `https://pixelcanvas.io/api/bigchunk/-10.5.bmp`  
    Make sure you disable the any caching for this request.

  - Result: raw image data, 4 bit per pixel:

    Buffer offset: (Has to be divided by 2 for higher and lower nibble)

    ``` PureBasic
    Offset = jx + jy * #Chunk_Size + (ix + iy * (#Chunk_Collection_Radius*2+1)) * #Chunk_Size * #Chunk_Size
    ```

    Chunk coordinates inside the chunk collection:

    ``` PureBasic
    CX = CCX * (#Chunk_Collection_Radius*2+1) + ix - #Chunk_Collection_Radius
    CY = CCY * (#Chunk_Collection_Radius*2+1) + iy - #Chunk_Collection_Radius
    ```

    `jx` and `jy` being the pixel-offset inside a chunk, and `ix` and `iy` being the chunk-offset inside a chunk collection.
    These counter variables must be greater or equal 0.
    `jx` and `jy` count from 0 to 63.
    `ix` and `iy` count from 0 to 14.
    Coordinate axis directions are pointing to the right and downwards.
    Even offsets are stored in the higher nibble.

- Get online player number:

  - URL: `https://pixelcanvas.io/api/online`

  - Result:JSON Object:

    | Key | Type | Description | Example |
    | --- | ---- | ----------- | ------- |
    | online            | int

- Get websocket URL:

  - URL: `https://pixelcanvas.io/api/ws`

  - Result:JSON Object:

    | Key | Type | Description | Example |
    | --- | ---- | ----------- | ------- |
    | url            | string

## Websocket protocol

### Connect to server

1. Make "Get websocket URL" HTTPS request
2. Connect to WS server: `URL + "/?fingerprint=" + fingerprint`

### Incoming

- Pixel Update

  Binary Frame:

  | Offset  | Length | Type | Description | Example |
  | ------  | ------ | ---- | ----------- | ------- |
  | 0       | 1     | int               | Opcode, always 0xC1
  | 1       | 2     | int big endian    | X chunk coordinate
  | 3       | 2     | int big endian    | Y chunk coordinate
  | 5       | 2     | int big endian    | `colorIndex = (number) & 0x0F` `offsetX = (number >> 4) & 0x3F` `offsetY = (number >> 10) & 0x3F`

## Tricks

### Get fingerprint from browser instance

``` JavaScript
webpackJsonp([0],{1000: (function(module, __webpack_exports__, __webpack_require__) {
    var __WEBPACK_IMPORTED_MODULE_1_fingerprintjs2__ = __webpack_require__(611);
    var __WEBPACK_IMPORTED_MODULE_1_fingerprintjs2___default = __webpack_require__.n(__WEBPACK_IMPORTED_MODULE_1_fingerprintjs2__);

    var fingerprint2 = new __WEBPACK_IMPORTED_MODULE_1_fingerprintjs2___default.a({
        extendedJsFonts: true
    });
    function getFingerprint() {
        return new Promise(function (resolve) {
            fingerprint2.get(resolve);
        });
    }

    console.log(getFingerprint());

})}, [1000]);
```

This is assuming that the fingerprint2 library was packed as module 611 with WebPack.
And that module 1000 isn't used already.