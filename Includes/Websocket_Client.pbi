; Websocketclient by Netzvamp / Robert Lieback
; Version: 2016/01/08
; Correction of that mess by David "D3" Vogel 2017-05-10

DeclareModule WebsocketClient
  Enumeration
    #Opcode_Continuation
    #Opcode_Text
    #Opcode_Binary
    
    #Opcode_Connection_Close = 8
    #Opcode_Ping
    #Opcode_Pong
  EndEnumeration
  
  #RSV1 = %00000100
  #RSV2 = %00000010
  #RSV3 = %00000001
  
  Declare   OpenWebsocketConnection(URL.s)
  
  Declare   Frame_Send(ConnectionID, *Data, Data_Size.i, Opcode.a=#Opcode_Binary)
  Declare   Frame_Text_Send(ConnectionID, Message.s)
  
  Declare   Frame_Receive(ConnectionID, *FrameType.Word=#Null)
  
  ;Declare   SetSSLProxy(ProxyServer.s = "", ProxyPort.l = 8182)
  
EndDeclareModule

Module WebsocketClient
  
  ;TODO: We don't support fragmetation right now
  ;TODO: We should send an closing frame, but server will also just close
  ;TODO: Support to send receive bigger frames
  
  Declare   Handshake(Connection, Servername.s, Path.s)
  Declare   ApplyMasking(Array Mask.a(1), *Data, Data_Size)
  
  Global Proxy_Server.s, Proxy_Port.l
  
  Macro dbg(txt)
    CompilerIf #PB_Compiler_Debugger
      Debug "WebsocketClient: " + FormatDate("%yyyy-%mm-%dd %hh:%ii:%ss",Date()) + " > " + txt
    CompilerEndIf
  EndMacro
  
  ;Procedure SetSSLProxy(ProxyServer.s = "", ProxyPort.l = 8182)
  ;  Proxy_Server.s = ProxyServer.s
  ;  Proxy_Port.l = ProxyPort.l
  ;EndProcedure
  
  Procedure OpenWebsocketConnection(URL.s)
    Protokol.s = GetURLPart(URL.s, #PB_URL_Protocol)
    Servername.s = GetURLPart(URL.s, #PB_URL_Site)
    Port.l = Val(GetURLPart(URL.s, #PB_URL_Port))
    If Port.l = 0 : Port.l = 80 : EndIf
    Path.s = GetURLPart(URL.s, #PB_URL_Path)
    If Path.s = "" : Path.s = "/" : EndIf
    Parameters.s = GetURLPart(URL, #PB_URL_Parameters)
    Address.s = Path
    If Parameters
      Address + "?" + Parameters
    EndIf
    
    InitNetwork()
    If Protokol.s = "wss" ; If we connect with encryption (https)
      If Proxy_Port
        Connection = OpenNetworkConnection(Proxy_Server.s, Proxy_Port.l, #PB_Network_TCP, 2000)
      Else
        dbg("We need an SSL-Proxy like stunnel for encryption. Configure the proxy with SetSSLProxy().")
      EndIf
    ElseIf Protokol.s = "ws"
      Connection = OpenNetworkConnection(Servername.s, Port.l, #PB_Network_TCP, 2000)
      ;Connection = Proxy::Proxy_Connect("5.9.145.114", 1117, Servername, Port, 1, "", "", 1000)
    EndIf
    
    If Connection
      If Handshake(Connection, Servername.s, Address.s)
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
  
  Procedure Handshake(Connection, Servername.s, Address.s)
    Request.s = "GET /" + Address.s + " HTTP/1.1"+ #CRLF$ +
                "Host: " + Servername.s + #CRLF$ +
                "Upgrade: websocket" + #CRLF$ +
                "Connection: Upgrade" + #CRLF$ +
                "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" + #CRLF$ +
                "Sec-WebSocket-Version: 13" + #CRLF$ + 
                "User-Agent: Mozilla/5.0 (Windows NT 10.0; WOW64; rv:53.0) Gecko/20100101 Firefox/53.0"+ #CRLF$ + #CRLF$
                
    Debug Request
                
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
    
    FreeMemory(*Buffer)
    
    ; Check answer
    Debug Answer
    If FindString(Answer.s, "HTTP/1.1 101") And FindString(Answer.s, "CONNECTION: UPGRADE") And FindString(Answer.s, "UPGRADE: WEBSOCKET")
      ProcedureReturn #True
    Else
      ProcedureReturn #False
    EndIf
  EndProcedure
  
  Procedure ApplyMasking(Array Mask.a(1), *Data, Data_Size)
    For i = 0 To Data_Size - 1
      PokeA(*Data + i, PeekA(*Data + i) ! Mask(i % 4))
    Next
  EndProcedure
  
  Procedure Frame_Send(ConnectionID, *Payload, Payload_Size.i, Opcode.a=#Opcode_Binary)
    Protected Transmit_Size
    Protected Transmit_Pos, Write_Pos
    Protected Result
    Protected Masking
    
    Transmit_Size = 2 + Payload_Size
    
    If Payload_Size > 0
      Masking = #True
      Transmit_Size + 4
    EndIf
    
    If Payload_Size <= 125
    ElseIf Payload_Size <= 65535
      Transmit_Size + 2 ; Additional 2 bytes for the payload size entry
    Else
      Transmit_Size + 8 ; Additional 8 bytes for the payload size entry
    EndIf
    
    *Buffer = AllocateMemory(Transmit_Size)
    
    ; #### Generate 4 random masking bytes
    If Masking
      Dim Mask.a(3)
      Mask(0) = Random(255,0)
      Mask(1) = Random(255,0)
      Mask(2) = Random(255,0)
      Mask(3) = Random(255,0)
    EndIf
    
    ; #### First Byte: FIN(1=finished with this Frame),RSV(0),RSV(0),RSV(0),OPCODE(4 byte)=0001(text) 
    PokeB(*Buffer + Write_Pos, %10000000 | (Opcode & %1111)) : Write_Pos + 1
    
    ; #### Second Byte: Masking(1),length(to 125bytes, else we have to extend)
    If Payload_Size <= 125                                                          ; Length fits in first byte
      PokeA(*Buffer + Write_Pos, Payload_Size | (Masking << 7))   : Write_Pos + 1   ; + 128 for Masking
    ElseIf Payload_Size <= 65535                                                    ; We have to extend length to third byte
      PokeA(*Buffer + Write_Pos, 126 | (Masking << 7))            : Write_Pos + 1   ; 126 for 2 extra length bytes and + 128 for Masking
      PokeA(*Buffer + Write_Pos, Payload_Size >> 8)               : Write_Pos + 1   ; First Byte
      PokeA(*Buffer + Write_Pos, Payload_Size)                    : Write_Pos + 1   ; Second Byte
    Else                                                                            ; It's bigger than 65535, we also use 8 extra bytes
      PokeA(*Buffer + Write_Pos, 127 | (Masking << 7))            : Write_Pos + 1   ; 127 for 8 extra length bytes and + 128 for Masking
      PokeA(*Buffer + Write_Pos, 0)                               : Write_Pos + 1   ; 8 Bytes for payload lenght. We don't support giant packages for now, so first bytes are zero :P
      PokeA(*Buffer + Write_Pos, 0)                               : Write_Pos + 1
      PokeA(*Buffer + Write_Pos, 0)                               : Write_Pos + 1
      PokeA(*Buffer + Write_Pos, 0)                               : Write_Pos + 1
      PokeA(*Buffer + Write_Pos, Payload_Size >> 24)              : Write_Pos + 1
      PokeA(*Buffer + Write_Pos, Payload_Size >> 16)              : Write_Pos + 1
      PokeA(*Buffer + Write_Pos, Payload_Size >> 8)               : Write_Pos + 1
      PokeA(*Buffer + Write_Pos, Payload_Size)                    : Write_Pos + 1   ; = 10 Byte
    EndIf
    
    ; #### Write Masking Bytes
    If Masking
      PokeA(*Buffer + Write_Pos, Mask(0))                         : Write_Pos + 1
      PokeA(*Buffer + Write_Pos, Mask(1))                         : Write_Pos + 1
      PokeA(*Buffer + Write_Pos, Mask(2))                         : Write_Pos + 1
      PokeA(*Buffer + Write_Pos, Mask(3))                         : Write_Pos + 1
    EndIf
    
    If *Payload And Payload_Size
      If Masking
        ApplyMasking(Mask(), *Payload, Payload_Size)
      EndIf
      CopyMemory(*Payload, *Buffer + Write_Pos, Payload_Size)     : Write_Pos + Payload_Size
    EndIf
    
    While Transmit_Pos < Transmit_Size
      Result = SendNetworkData(ConnectionID, *Buffer + Transmit_Pos, Transmit_Size - Transmit_Pos)
      If Result >= 0
        Transmit_Pos + Result
      Else
        FreeMemory(*Buffer)
        ProcedureReturn #False
      EndIf
    Wend
    
    FreeMemory(*Buffer)
    ProcedureReturn #True
  EndProcedure
  
  Procedure Frame_Text_Send(ConnectionID, Message.s)
    
    Protected *Message = UTF8(Message)
    Message_Size = MemorySize(*Message)
    
    Result = Frame_Send(ConnectionID, *Message, Message_Size, #Opcode_Text)
    
    FreeMemory(*Message)
    ProcedureReturn Result
  EndProcedure
  
  Procedure Frame_Receive(ConnectionID, *FrameType.Word=#Null)
    Protected *Buffer = AllocateMemory(10)
    Protected *Temp
    Protected Receive_Pos = 0
    Protected Receive_Size = 2
    Protected Read_Pos = 0
    Protected Temp_Result
    Protected *Message
    Protected Fragmented
    Protected Payload_Size
    
    While Receive_Pos < Receive_Size
      Temp_Result = ReceiveNetworkData(ConnectionID, *Buffer + Receive_Pos, Receive_Size - Receive_Pos)
      If Temp_Result < 0
        FreeMemory(*Buffer)
        ProcedureReturn #Null
      Else
        Receive_Pos + Temp_Result
      EndIf
    Wend
    
    ; #### Getting informations about packet
    If PeekA(*Buffer) & %10000000
      Fragmented = #False
    Else
      Fragmented = #True
    EndIf
    
    ; #### Check for Opcodes
    If *FrameType : *FrameType\w = PeekA(*Buffer) & %1111 : EndIf
    Select PeekA(*Buffer) & %1111
      Case #Opcode_Continuation
      Case #Opcode_Text
      Case #Opcode_Binary
      Case #Opcode_Connection_Close
      Case #Opcode_Ping
        Frame_Send(ConnectionID, #Null, 0, #Opcode_Pong)
        FreeMemory(*Buffer)
        ProcedureReturn #Null
      Case #Opcode_Pong
      Default
        FreeMemory(*Buffer)
        ProcedureReturn #Null
    EndSelect
    
    Read_Pos + 1
    
    ; #### Check masking
    If PeekA(*Buffer + Read_Pos) & %10000000
      Masking = #True
      Receive_Size + 4
    Else
      Masking = #False
    EndIf
    
    ; #### Check size
    If PeekA(*Buffer + Read_Pos) & %01111111 <= 125 ; size is in this byte
      Payload_Size = PeekA(*Buffer + Read_Pos) & %01111111 : Read_Pos + 1
    ElseIf PeekA(*Buffer + Read_Pos) & %01111111 >= 126 ; Size is in 2 extra bytes
      
      ; #### Receive 2 additional bytes
      Receive_Size + 2
      While Receive_Pos < Receive_Size
        Temp_Result = ReceiveNetworkData(ConnectionID, *Buffer + Receive_Pos, Receive_Size - Receive_Pos)
        If Temp_Result < 0
          FreeMemory(*Buffer)
          ProcedureReturn #Null
        Else
          Receive_Pos + Temp_Result
        EndIf
      Wend
      
      Read_Pos + 1
      
      Payload_Size = PeekA(*Buffer + Read_Pos) << 8 + PeekA(*Buffer + Read_Pos + 1) : Read_Pos + 2
    Else
      ; TODO: Ability to receive large frames
      ProcedureReturn #Null
    EndIf
    
    Receive_Size + Payload_Size
    
    ; #### Resize buffer
    *Temp = ReAllocateMemory(*Buffer, Receive_Size)
    If *Temp
      *Buffer = *Temp
    Else
      FreeMemory(*Buffer)
      ProcedureReturn #Null
    EndIf
    
    ; #### Receive the main data
    While Receive_Pos < Receive_Size
      Temp_Result = ReceiveNetworkData(ConnectionID, *Buffer + Receive_Pos, Receive_Size - Receive_Pos)
      If Temp_Result < 0
        FreeMemory(*Buffer)
        ProcedureReturn #Null
      Else
        Receive_Pos + Temp_Result
      EndIf
    Wend
    
    If Masking = #True
      Dim Mask.a(3)
      Mask(0) = PeekA(*Buffer + Read_Pos) : Read_Pos + 1
      Mask(1) = PeekA(*Buffer + Read_Pos) : Read_Pos + 1
      Mask(2) = PeekA(*Buffer + Read_Pos) : Read_Pos + 1
      Mask(3) = PeekA(*Buffer + Read_Pos) : Read_Pos + 1
      
      *Message = AllocateMemory(Receive_Size - Read_Pos)
      CopyMemory(*Buffer + Read_Pos, *Message, Receive_Size - Read_Pos)
      
      ApplyMasking(Mask(), *Message, Receive_Size - Read_Pos)
    ElseIf Receive_Size - Read_Pos > 0
      *Message = AllocateMemory(Receive_Size - Read_Pos)
      CopyMemory(*Buffer + Read_Pos, *Message, Receive_Size - Read_Pos)
    EndIf
    
    FreeMemory(*Buffer)
    ProcedureReturn *Message
  EndProcedure
  
EndModule
; IDE Options = PureBasic 5.60 (Windows - x64)
; CursorPosition = 324
; FirstLine = 305
; Folding = --
; EnableXP
; DisableDebugger
; EnableCompileCount = 3
; EnableBuildCount = 0
; EnableExeConstant