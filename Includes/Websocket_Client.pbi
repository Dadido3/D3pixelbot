; Websocketclient by Netzvamp
; Version: 2016/01/08

DeclareModule WebsocketClient
  Declare OpenWebsocketConnection(URL.s)
  Declare SendTextFrame(connection, message.s)
  Declare ReceiveFrame(connection, *MsgBuffer)
  Declare SetSSLProxy(ProxyServer.s = "", ProxyPort.l = 8182)
  
  Enumeration
    #frame_text
    #frame_binary
    #frame_closing
    #frame_ping
    #frame_unknown
  EndEnumeration
  
EndDeclareModule

Module WebsocketClient
  
  ;TODO: Add function to send binary frame
  ;TODO: We don't support fragmetation right now
  ;TODO: We should send an closing frame, but server will also just close
  ;TODO: Support to send receive bigger frames
  
  Declare Handshake(Connection, Servername.s, Path.s)
  Declare ApplyMasking(Array Mask.a(1), *Buffer)
  
  Global Proxy_Server.s, Proxy_Port.l
  
  Macro dbg(txt)
    CompilerIf #PB_Compiler_Debugger
      Debug "WebsocketClient: " + FormatDate("%yyyy-%mm-%dd %hh:%ii:%ss",Date()) + " > " + txt
    CompilerEndIf
  EndMacro
  
  Procedure SetSSLProxy(ProxyServer.s = "", ProxyPort.l = 8182)
    Proxy_Server.s = ProxyServer.s
    Proxy_Port.l = ProxyPort.l
  EndProcedure
  
  Procedure OpenWebsocketConnection(URL.s)
    Protokol.s = GetURLPart(URL.s, #PB_URL_Protocol)
    Servername.s = GetURLPart(URL.s, #PB_URL_Site)
    Port.l = Val(GetURLPart(URL.s, #PB_URL_Port))
    If Port.l = 0 : Port.l = 80 : EndIf
    Path.s = GetURLPart(URL.s, #PB_URL_Path)
    If Path.s = "" : Path.s = "/" : EndIf
    
    InitNetwork()
    If Protokol.s = "wss" ; If we connect with encryption (https)
      If Proxy_Port
        Connection = OpenNetworkConnection(Proxy_Server.s, Proxy_Port.l, #PB_Network_TCP, 1000)
      Else
        dbg("We need an SSL-Proxy like stunnel for encryption. Configure the proxy with SetSSLProxy().")
      EndIf
    ElseIf Protokol.s = "ws"
      Connection = OpenNetworkConnection(Servername.s, Port.l, #PB_Network_TCP, 1000)
    EndIf
    
    If Connection
      If Handshake(Connection, Servername.s, Path.s)
        dbg("Connection and Handshake ok")
        ProcedureReturn Connection
      Else
        dbg("Handshake-Error")
        ProcedureReturn #False
      EndIf
    Else
      dbg("Couldn't connect")
      ProcedureReturn #False
    EndIf
  EndProcedure
  
  Procedure Handshake(Connection, Servername.s, Path.s)
    Request.s = "GET /" + Path.s + " HTTP/1.1"+ #CRLF$ +
                "Host: " + Servername.s + #CRLF$ +
                "Upgrade: websocket" + #CRLF$ +
                "Connection: Upgrade" + #CRLF$ +
                "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" + #CRLF$ +
                "Sec-WebSocket-Version: 13" + #CRLF$ + 
                "User-Agent: CustomWebsocketClient"+ #CRLF$ + #CRLF$
                
    SendNetworkString(Connection, Request.s, #PB_UTF8)
    *Buffer = AllocateMemory(65536)
    
    ; We wait for answer
    Repeat
      Size = ReceiveNetworkData(connection, *Buffer, 65536)
      Answer.s = Answer.s + PeekS(*Buffer, Size, #PB_UTF8)
      If FindString(Answer, #CRLF$ + #CRLF$)
        Break
      EndIf
    Until Size <> 65536
    
    Answer.s = UCase(Answer.s)
    
    ; Check answer
    If FindString(Answer.s, "HTTP/1.1 101") And FindString(Answer.s, "CONNECTION: UPGRADE") And FindString(Answer.s, "UPGRADE: WEBSOCKET")
      ProcedureReturn #True
    Else
      ProcedureReturn #False
    EndIf
  EndProcedure
  
  Procedure ApplyMasking(Array Mask.a(1), *Buffer)
    For i = 0 To MemorySize(*Buffer) - 1
      PokeA(*Buffer + i, PeekA(*Buffer + i) ! Mask(i % 4))
    Next
  EndProcedure
  
  Procedure SendTextFrame(connection, message.s)
    
    ; Put String in Buffer
    MsgLength.l = StringByteLength(message.s, #PB_UTF8)
    *MsgBuffer = AllocateMemory(MsgLength)
    PokeS(*MsgBuffer, message.s, MsgLength, #PB_UTF8|#PB_String_NoZero)
    
    dbg("Messagelength to send: " + Str(MsgLength))
    
    ; The Framebuffer, we fill with senddata
    If MsgLength <= 125
      Fieldlength = 6
    ElseIf MsgLength >= 126 And MsgLength <= 65535
      Fieldlength = 8
    Else
      Fieldlength = 14
    EndIf
    
    dbg("Fieldlength to send: " + Str(Fieldlength))
    
    
    *FrameBuffer = AllocateMemory(Fieldlength + MsgLength)
    
    ; We generate 4 random masking bytes
    Dim Mask.a(3)
    Mask(0) = Random(255,0)
    Mask(1) = Random(255,0) 
    Mask(2) = Random(255,0) 
    Mask(3) = Random(255,0) 
    
    pos = 0 ; The byteposotion in the framebuffer
    
    ; First Byte: FIN(1=finished with this Frame),RSV(0),RSV(0),RSV(0),OPCODE(4 byte)=0001(text) 
    PokeB(*FrameBuffer, %10000001) : pos + 1 ; = 129
    
    ; Second Byte: Masking(1),length(to 125bytes, else we have to extend)
    If MsgLength <= 125                                             ; Length fits in first byte
      PokeA(*Framebuffer + pos, MsgLength + 128)    : pos + 1       ; + 128 for Masking
    ElseIf MsgLength >= 126 And MsgLength <= 65535                  ; We have to extend length to third byte
      PokeA(*Framebuffer + pos, 126 + 128)          : pos + 1       ; 126 for 2 extra length bytes and + 128 for Masking
      PokeA(*FrameBuffer + pos, (MsgLength >> 8))   : pos + 1       ; First Byte
      PokeA(*FrameBuffer + pos, MsgLength)          : pos + 1       ; Second Byte
    Else                                                            ; It's bigger than 65535, we also use 8 extra bytes
      PokeA(*Framebuffer + pos, 127 + 128)          : pos + 1       ; 127 for 8 extra length bytes and + 128 for Masking
      PokeA(*Framebuffer + pos, 0)                  : pos + 1       ; 8 Bytes for payload lenght. We don't support giant packages for now, so first bytes are zero :P
      PokeA(*Framebuffer + pos, 0)                  : pos + 1
      PokeA(*Framebuffer + pos, 0)                  : pos + 1
      PokeA(*Framebuffer + pos, 0)                  : pos + 1
      PokeA(*Framebuffer + pos, MsgLength >> 24)    : pos + 1
      PokeA(*Framebuffer + pos, MsgLength >> 16)    : pos + 1
      PokeA(*Framebuffer + pos, MsgLength >> 8)     : pos + 1
      PokeA(*Framebuffer + pos, MsgLength)          : pos + 1       ; = 10 Byte
    EndIf
    ; Write Masking Bytes
    PokeA(*FrameBuffer + pos, Mask(0))              : pos + 1
    PokeA(*FrameBuffer + pos, Mask(1))              : pos + 1
    PokeA(*FrameBuffer + pos, Mask(2))              : pos + 1
    PokeA(*FrameBuffer + pos, Mask(3))              : pos + 1
    
    ApplyMasking(Mask(), *MsgBuffer)
    
    CopyMemory(*MsgBuffer, *FrameBuffer + pos, MsgLength)
    
    ;For x = 0 To 100 Step 5
      ;Debug Str(PeekA(*FrameBuffer + x)) + " | " + Str(PeekA(*FrameBuffer + x + 1)) + " | " + Str(PeekA(*FrameBuffer + x + 2)) + " | " + Str(PeekA(*FrameBuffer + x + 3)) + " | " + Str(PeekA(*FrameBuffer + x + 4))
    ;Next
    
    If SendNetworkData(connection, *FrameBuffer, Fieldlength + MsgLength) = Fieldlength + MsgLength
      dbg("Textframe send, Bytes: " + Str(Fieldlength + MsgLength))
      ProcedureReturn #True
    Else
      ProcedureReturn #False
    EndIf
    
  EndProcedure
  
  Procedure ReceiveFrame(connection, *MsgBuffer)
    
    *FrameBuffer = AllocateMemory(65536)
    
    Repeat
      *FrameBuffer = ReAllocateMemory(*FrameBuffer, 65536)
      Size = ReceiveNetworkData(connection, *FrameBuffer, 65536)
      ;Answer.s = Answer.s + PeekS(*FrameBuffer, Size, #PB_UTF8)
    Until Size <> 65536
    
    dbg("Received Frame, Bytes: " + Str(Size))
    
    *FrameBuffer = ReAllocateMemory(*FrameBuffer, Size)
    
    ;     ; debug: output any single byte
    ;     If #PB_Compiler_Debugger
    ;       For x = 0 To Size - 1 Step 1
    ;         dbg_bytes.s + Str(PeekA(*FrameBuffer + x)) + " | "
    ;       Next
    ;       dbg(dbg_bytes)
    ;     EndIf
    
    ; Getting informations about package
    If PeekA(*FrameBuffer) & %10000000 > #False
      ;dbg("Frame not fragmented")
      fragmentation.b = #False
    Else
      dbg("Frame fragmented! This not supported for now!")
      fragmentation.b = #True
    EndIf
    
    ; Check for Opcodes
    If PeekA(*FrameBuffer) = %10000001 ; Textframe
      dbg("Text frame")
      frame_typ.w = #frame_text
    ElseIf PeekA(*FrameBuffer) = %10000010 ; Binary Frame
      dbg("Binary frame")
      frame_typ.w = #frame_binary
    ElseIf PeekA(*FrameBuffer) = %10001000 ; Closing Frame
      dbg("Closing frame")
      frame_typ.w = #frame_closing
    ElseIf PeekA(*FrameBuffer) = %10001001 ; Ping
      ; We just answer pings
      *pongbuffer = AllocateMemory(2)
      PokeA(*pongbuffer, 138)
      PokeA(*pongbuffer+1, 0)
      SendNetworkData(connection, *pongbuffer, 2)
      dbg("Received Ping, answered with Pong")
      frame_typ.w = #frame_ping
      ProcedureReturn
    Else
      dbg("Opcode unknown")
      frame_typ.w = #frame_unknown
      ProcedureReturn #False
    EndIf
    
    ; Check masking
    If PeekA(*FrameBuffer + 1) & %10000000 = 128 : masking.b = #True : Else : masking.b = #False : EndIf
    
    dbg("Masking: " + Str(masking))
    
    pos.l = 1
    
    ; check size
    If PeekA(*FrameBuffer + 1) & %01111111 <= 125 ; size is in this byte
      frame_size.l = PeekA(*FrameBuffer + pos) & %01111111 : pos + 1
    ElseIf PeekA(*FrameBuffer + 1) & %01111111 >= 126 ; Size is in 2 extra bytes
      frame_size.l = PeekA(*FrameBuffer + 2) << 8 + PeekA(*FrameBuffer + 3) : pos + 2
    EndIf
    dbg("FrameSize: " + Str(frame_size.l))
    
    If masking = #True
      Dim Mask.a(3)
      Mask(0) = PeekA(*FrameBuffer + pos) : pos + 1
      Mask(1) = PeekA(*FrameBuffer + pos) : pos + 1
      Mask(2) = PeekA(*FrameBuffer + pos) : pos + 1
      Mask(3) = PeekA(*FrameBuffer + pos) : pos + 1
      
      ReAllocateMemory(*MsgBuffer,frame_size)
      CopyMemory(*FrameBuffer + pos, *MsgBuffer, frame_size)
      
      ApplyMasking(Mask(), *MsgBuffer)
    Else
      ReAllocateMemory(*MsgBuffer,frame_size)
      CopyMemory(*FrameBuffer + pos, *MsgBuffer, frame_size)
    EndIf
    
    ProcedureReturn frame_typ
    
  EndProcedure
  
EndModule


CompilerIf #PB_Compiler_IsMainFile
  
  ; Minimal example to send and receive textmessages
  ; The preconfigured testserver "echo.websocket.org" will just echo back everything you've send.
  
  XIncludeFile "module_websocketclient_gui.pbf"
  
  Global connection
  
  Procedure gui_button_connect(EventType)
    If EventType = #PB_EventType_LeftClick
      Debug "Connect clicked"
      ;connection = WebsocketClient::OpenWebsocketConnection(GetGadgetText(#String_url))
      connection = WebsocketClient::OpenWebsocketConnection(GetGadgetText(#String_url))
      
      AddGadgetItem(#ListView_Output, -1, "# Connected to " + GetGadgetText(#String_url))
    EndIf
  EndProcedure
  
  Procedure gui_button_send(EventType)
    If EventType = #PB_EventType_LeftClick And Len(GetGadgetText(#String_send)) > 0 And connection
      Debug "Send clicked"
      If WebsocketClient::SendTextFrame(connection, GetGadgetText(#String_send)) = #False
        Debug "Couldn't send. Are we disconnected?"
      Else
        AddGadgetItem(#ListView_Output, -1, "> " + GetGadgetText(#String_send))
      EndIf
    EndIf
  EndProcedure
  
  OpenWindow_Websocketclient()
  
  ; Proxy Setting:
  ; If you need an encyrpted connection (https/wss), you currently have to use an 
  ; proxy software like stunnel (https://www.stunnel.org) to redirect unencrypted data into an encrypted connection
  ; Example stunnel.conf section:
  ;   [websocket]
  ;   client = yes
  ;   accept = 127.0.0.1:8182
  ;   connect = echo.websocket.org:443
  WebsocketClient::SetSSLProxy("127.0.0.1",8182)
  
  Repeat
    
    If Window_Websocketclient_Events( WaitWindowEvent(1) ) = #False : End : EndIf
    
    If connection
      
      NetworkEvent = NetworkClientEvent(connection)
      
      Select NetworkEvent
          
        Case #PB_NetworkEvent_Data
          Debug "We've got Data"
          *FrameBuffer = AllocateMemory(1)
          Frametyp = WebsocketClient::ReceiveFrame(connection,*FrameBuffer)
          If Frametyp = WebsocketClient::#frame_text
            AddGadgetItem(#ListView_Output, -1, "< " + PeekS(*FrameBuffer,MemoryStringLength(*FrameBuffer,#PB_UTF8)-1,#PB_UTF8) )
          ElseIf Frametyp = WebsocketClient::#frame_binary
            AddGadgetItem(#ListView_Output, -1, "< Received Binaryframe" )
          EndIf
          
        Case #PB_NetworkEvent_Disconnect
          If disconnected = #False
            Debug "Disconnected"
          EndIf
          disconnected = #True
          NetworkEvent = #PB_NetworkEvent_None
          
        Case #PB_NetworkEvent_None
          
      EndSelect
      
    EndIf
  ForEver
  
CompilerEndIf