; ##################################################### License / Copyright #########################################
; 
; ##################################################### Documentation / Comments ####################################
; 
; History:
; - V0.000 (09.05.2017)
;   - Project started
; 
; ##################################################### External Includes ###########################################
XIncludeFile "Includes/Crash.pbi"
XIncludeFile "Includes/Helper.pbi"
;XIncludeFile "Includes/Proxy.pbi"
XIncludeFile "Includes/Websocket_Client.pbi"

; ##################################################### Includes ####################################################

DeclareModule Main
  EnableExplicit
  
  ; ################################################### Prototypes ##################################################
  
  ; ################################################### Constants ###################################################
  #Version = 0953
  
  #Software_Name = "Pixelcanvas.io Custom Client"
  
  #WebSocket_Reconnect_Time = 10000
  
  #Chunk_Size = 64 ; Pixel
  #Chunk_Collection_Radius = 7
  
  #Chunk_Collection_Download_Timeout = 5 * 60 * 1000
  
  #Colors = 16
  
  #Filename_Settings = "Settings.txt"
  #Filename_Templates = "Templates.txt"
  
  Enumeration 1
    #Menu_Dummy
    
    #Menu_Templates
    
    #Menu_Canvas_Load
    #Menu_Canvas_Reload
    #Menu_Canvas_AutoReload
    
    #Menu_Captcha_Requester
    
    #Menu_Settings_Change_Fingerprint
    
    #Menu_About
    
    #Menu_Exit
  EndEnumeration
  
  Enumeration
    #Input_Result_Success
    
    #Input_Result_Local_Error
    #Input_Result_Global_Error
  EndEnumeration
  
  ; ################################################### Structures ##################################################
  Structure Main
    Quit.i
    
    ;Timestamp_Offset.q            ; Server timestamp offset (to ElapsedMilliseconds())
    ;Timestamp_Next_TimeRequest.q  ; Point of time, when to make a new timestamp request (ElapsedMilliseconds())
    ;Timestamp_Counter.u           ; Counter for the timestamp request
    
    Timer_PPS.q
    Counter_PPS.i
    PPS.d
    
    Path_AppData.s
  EndStructure
  
  Structure Userdata
    Logged_In.i
    
    ID.s
    Name.s
    Region.s
    
    Center_X.i
    Center_Y.i
    
    Timestamp_Next_Pixel.q        ; Point of time, when the next pixel is drawable (Get_Timestamp())
  EndStructure
  
  Structure Window_Canvas
    ID.i
    Redraw.i
    
    Show_Templates_Blink_Timer.i
    Show_Templates.i
    
    Mouse_Translate.i
    Mouse_Translate_X.i
    Mouse_Translate_Y.i
  EndStructure
  
  Structure Window
    ID.i
    
    Menu_ID.i
    Menu_Height.i
    
    ToolBar.i
    ToolBar_Height.i
    
    StatusBar.i
    StatusBar_Height.i
    
    Canvas.Window_Canvas
  EndStructure
  
  Structure Chunk
    CX.i            ; Absolute chunk position index
    CY.i            ; Absolute chunk Position index
    
    Image.i
    Image_Scaled.i
    Image_Scale.d   ; Current scale factor of the scaled image
    
    Update.i
    Visible.i
  EndStructure
  
  Structure Chunk_Collection  ; Collection of chunks, center is at CCX * #Chunk_Collection_Radius * #Chunk_Size, CCY * ...
    CCX.i                     ; Collection position index (Center)
    CCY.i                     ; Collection position index (Center)
    
    Downloaded.i              ; #False: Is being downloaded, #True: Downloaded
    Asynchronous_Download.i   ; The connection for the asynchronous download
    Download_Timeout.q
    
    List Chunk.Chunk()
  EndStructure
  
  Structure RGBA_Struct
    R.a
    G.a
    B.a
    A.a
    
    Color.l
  EndStructure
  
  Structure WebSocket
    Connection.i
    
    Reconnect_Timer.i
    
    URL.s
  EndStructure
  
  Structure Input_Blocking ; List of input events, which are blocked due to negative response from the server
    X.w
    Y.w
    Color_Index.a
    
    Timestamp.q
  EndStructure
  
  Structure Settings
    Canvas_AutoReload.i
    Captcha_Requester.i
    
    X.d
    Y.d
    Zoom.d
    
    Fingerprint.s
  EndStructure
  
  ; ################################################### Variables ###################################################
  Global Main.Main
  Global Userdata.Userdata
  Global Window.Window
  Global WebSocket.WebSocket
  Global Settings.Settings
  
  Global NewList Chunk_Collection.Chunk_Collection()
  Global NewList Input_Blocking.Input_Blocking()
  
  Global Dim Palette.RGBA_Struct(#Colors-1)
  
  ; ################################################### Macros ######################################################
  
  ; ################################################### Declares ####################################################
  Declare.q Get_Timestamp()
  Declare   Get_Color_Index(Color)
  
  Declare   Image_Get(X.i, Y.i, Width.i, Height.i)
  
  Declare   Chunk_Collection_Get(CCX.i, CCY.i, Create=#False)
  Declare   Chunk_Collection_Download_Area(X.i, Y.i, Width.i, Height.i)
  Declare   Chunk_Collection_Delete(*Chunk_Collection.Chunk_Collection)
  Declare   _Chunk_Collection_Create(CCX.i, CCY.i)
  
  Declare   WebSocket_Send_Input(X.w, Y.w, Color_Index.a, *Template=#Null)
  Declare   HTTP_Post_Input(X, Y, Color_Index.a, *Template=#Null, Fingerprint.s="")
  
EndDeclareModule

; ##################################################### Includes ####################################################
XIncludeFile "Includes/About.pbi"
XIncludeFile "Includes/Captcha_Requester.pbi"
XIncludeFile "Includes/Templates.pbi"

Module Main
  EnableExplicit
  
  UseModule Helper
  
  InitNetwork()
  
  UsePNGImageDecoder()
  UseMD5Fingerprint()
  
  ; ################################################### Init ########################################################
  Global Font_Normal = LoadFont(#PB_Any, "Arial", 10)
  
  Palette(00)\R = 255 : Palette(00)\G = 255 : Palette(00)\B = 255 : Palette(00)\A = 255
  Palette(01)\R = 228 : Palette(01)\G = 228 : Palette(01)\B = 228 : Palette(01)\A = 255
  Palette(02)\R = 136 : Palette(02)\G = 136 : Palette(02)\B = 136 : Palette(02)\A = 255
  Palette(03)\R = 034 : Palette(03)\G = 034 : Palette(03)\B = 034 : Palette(03)\A = 255
  Palette(04)\R = 255 : Palette(04)\G = 167 : Palette(04)\B = 209 : Palette(04)\A = 255
  Palette(05)\R = 229 : Palette(05)\G = 000 : Palette(05)\B = 000 : Palette(05)\A = 255
  Palette(06)\R = 229 : Palette(06)\G = 149 : Palette(06)\B = 000 : Palette(06)\A = 255
  Palette(07)\R = 160 : Palette(07)\G = 106 : Palette(07)\B = 066 : Palette(07)\A = 255
  Palette(08)\R = 229 : Palette(08)\G = 217 : Palette(08)\B = 000 : Palette(08)\A = 255
  Palette(09)\R = 148 : Palette(09)\G = 224 : Palette(09)\B = 068 : Palette(09)\A = 255
  Palette(10)\R = 002 : Palette(10)\G = 190 : Palette(10)\B = 001 : Palette(10)\A = 255
  Palette(11)\R = 000 : Palette(11)\G = 211 : Palette(11)\B = 221 : Palette(11)\A = 255
  Palette(12)\R = 000 : Palette(12)\G = 131 : Palette(12)\B = 199 : Palette(12)\A = 255
  Palette(13)\R = 000 : Palette(13)\G = 000 : Palette(13)\B = 234 : Palette(13)\A = 255
  Palette(14)\R = 207 : Palette(14)\G = 110 : Palette(14)\B = 228 : Palette(14)\A = 255
  Palette(15)\R = 130 : Palette(15)\G = 000 : Palette(15)\B = 128 : Palette(15)\A = 255
  
  Define i
  For i = 0 To #Colors-1
    Palette(i)\Color = RGBA(Palette(i)\R, Palette(i)\G, Palette(i)\B, Palette(i)\A)
  Next
  
  ; ################################################### Declares ####################################################
  Declare   Main()
  
  ; ################################################### Icons ... ###################################################
  Global Icon_images = CatchImage(#PB_Any, ?Icon_images)
  Global Icon_map = CatchImage(#PB_Any, ?Icon_map)
  Global Icon_map_go = CatchImage(#PB_Any, ?Icon_map_go)
  Global Icon_time_go = CatchImage(#PB_Any, ?Icon_time_go)
  Global Icon_key = CatchImage(#PB_Any, ?Icon_key)
  Global Icon_information = CatchImage(#PB_Any, ?Icon_information)
  Global Icon_bell = CatchImage(#PB_Any, ?Icon_bell)
  
  
  ; ################################################### Regular Expressions #########################################
  Global RegEx_Duck = CreateRegularExpression(#PB_Any, "DUCK=(?<Duck>[a-z])")
  
  ; ################################################### Procedures ##################################################
  Procedure Settings_Save(Filename.s)
    Protected JSON = CreateJSON(#PB_Any)
    If JSON
      InsertJSONStructure(JSONValue(JSON), Settings, Settings)
      SaveJSON(JSON, Filename, #PB_JSON_PrettyPrint)
      FreeJSON(JSON)
    EndIf
  EndProcedure
  
  Procedure Settings_Load(Filename.s)
    Protected JSON = LoadJSON(#PB_Any, Filename)
    If JSON
      ExtractJSONStructure(JSONValue(JSON), Settings, Settings)
      FreeJSON(JSON)
    EndIf
    
    If Settings\Zoom = 0
      Settings\Zoom = 1
    EndIf
    
    If Settings\Fingerprint = ""
      Define Random.q = Random(2147483647)
      Settings\Fingerprint = Fingerprint(@Random, 8, #PB_Cipher_MD5)
    EndIf
    
  EndProcedure
  
  Procedure.q Get_Timestamp()
    ProcedureReturn ElapsedMilliseconds(); + Main\Timestamp_Offset
  EndProcedure
  
  Procedure Get_Color_Index(Color)
    Protected i
    Protected Color_Index
    Protected Distance.d, Temp_Distance.d
    
    For i = 0 To #Colors-1
      Distance.d = Sqr(Pow(Palette(i)\R - Red(Color),2) + Pow(Palette(i)\G - Green(Color),2) + Pow(Palette(i)\B - Blue(Color),2))
      If i = 0 Or Temp_Distance > Distance
        Temp_Distance = Distance
        Color_Index = i
      EndIf
    Next
    
    ProcedureReturn Color_Index
  EndProcedure
  
  Procedure Canvas_Event()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    Protected Mouse_X = GetGadgetAttribute(Event_Gadget, #PB_Canvas_MouseX)
    Protected Mouse_Y = GetGadgetAttribute(Event_Gadget, #PB_Canvas_MouseY)
    
    Protected Temp_Zoom.d
    
    Protected Width = GadgetWidth(Event_Gadget)
    Protected Height = GadgetHeight(Event_Gadget)
    
    Select Event_Type
      Case #PB_EventType_MiddleButtonDown, #PB_EventType_RightButtonDown
        Window\Canvas\Mouse_Translate = #True
        Window\Canvas\Mouse_Translate_X = Mouse_X
        Window\Canvas\Mouse_Translate_Y = Mouse_Y
        
      Case #PB_EventType_MiddleButtonUp, #PB_EventType_RightButtonUp
        Window\Canvas\Mouse_Translate = #False
        
      Case #PB_EventType_MouseMove
        If Window\Canvas\Mouse_Translate
          If Window\Canvas\Mouse_Translate_X - Mouse_X Or Window\Canvas\Mouse_Translate_Y - Mouse_Y
            Settings\X + Mouse_X - Window\Canvas\Mouse_Translate_X
            Settings\Y + Mouse_Y - Window\Canvas\Mouse_Translate_Y
            Window\Canvas\Mouse_Translate_X = Mouse_X
            Window\Canvas\Mouse_Translate_Y = Mouse_Y
            Window\Canvas\Redraw = #True
            
          EndIf
        EndIf
        Protected X_R.d = (Mouse_X - Width / 2 - Settings\X) / Settings\Zoom
        Protected Y_R.d = (Mouse_Y - Height / 2 - Settings\Y) / Settings\Zoom
        StatusBarText(Window\StatusBar, 1, "X: " + Str(Round(X_R, #PB_Round_Down)), #PB_StatusBar_Right)
        StatusBarText(Window\StatusBar, 2, "Y: " + Str(Round(Y_R, #PB_Round_Down)), #PB_StatusBar_Right)
        
      Case #PB_EventType_MouseWheel
        Temp_Zoom = Pow(2, GetGadgetAttribute(Event_Gadget, #PB_Canvas_WheelDelta))
        If Settings\Zoom * Temp_Zoom < Pow(2, -2)
          Temp_Zoom = Pow(2, -2) / Settings\Zoom
        EndIf
        If Settings\Zoom * Temp_Zoom > Pow(2, 4)
          Temp_Zoom = Pow(2, 4) / Settings\Zoom
        EndIf
        Settings\X - (Temp_Zoom - 1) * (GetGadgetAttribute(Event_Gadget, #PB_Canvas_MouseX) - Settings\X - Width/2)
        Settings\Y - (Temp_Zoom - 1) * (GetGadgetAttribute(Event_Gadget, #PB_Canvas_MouseY) - Settings\Y - Height/2)
        Settings\Zoom * Temp_Zoom
        
        Window\Canvas\Redraw = #True
        
      Case #PB_EventType_Input
        
      Case #PB_EventType_Focus
        
      Case #PB_EventType_KeyDown
        
    EndSelect
    
  EndProcedure
  
  Procedure Window_StatusBar_Update()
    StatusBarText(Window\StatusBar, 0, StrD((Userdata\Timestamp_Next_Pixel - Get_Timestamp()) / 1000, 1) + "s", #PB_StatusBar_Right)
    StatusBarText(Window\StatusBar, 4, Str(ListSize(Input_Blocking())) + " inputs failed", #PB_StatusBar_Right)
  EndProcedure
  
  Procedure Window_Event_SizeWindow()
    Protected Event_Window = EventWindow()
    
    Protected Width = WindowWidth(Event_Window)
    Protected Height = WindowHeight(Event_Window)
    
    ResizeGadget(Window\Canvas\ID, 0, Window\ToolBar_Height, Width, Height-Window\Menu_Height-Window\StatusBar_Height-Window\ToolBar_Height)
    
    Window\Canvas\Redraw = #True
  EndProcedure
  
  Procedure Window_Event_Menu()
    Protected Event_Menu = EventMenu()
    
    Select Event_Menu
      Case #Menu_Templates
        Templates::Window_Open()
        
      Case #Menu_Canvas_Load
        Protected Width = GadgetWidth(Window\Canvas\ID)
        Protected Height = GadgetHeight(Window\Canvas\ID)
        Protected X_R_1.d, Y_R_1.d, X_R_2.d, Y_R_2.d
        X_R_1 = (-Width / 2 - Settings\X) / Settings\Zoom
        Y_R_1 = (-Height / 2 - Settings\Y) / Settings\Zoom
        X_R_2 = (Width / 2 - Settings\X) / Settings\Zoom
        Y_R_2 = (Height / 2 - Settings\Y) / Settings\Zoom
        ; #### Load chunks
        Chunk_Collection_Download_Area(X_R_1, Y_R_1, X_R_2-X_R_1, Y_R_2-Y_R_1)
        
      Case #Menu_Canvas_Reload
        ForEach Chunk_Collection()
          Chunk_Collection()\Downloaded = #False
        Next
        
      Case #Menu_Canvas_AutoReload
        Settings\Canvas_AutoReload = GetToolBarButtonState(Window\ToolBar, #Menu_Canvas_AutoReload)
        
      Case #Menu_Captcha_Requester
        Settings\Captcha_Requester = GetToolBarButtonState(Window\ToolBar, #Menu_Captcha_Requester)
        
      Case #Menu_Settings_Change_Fingerprint
        Settings\Fingerprint = InputRequester("Change fingerprint", "Enter the new fingerprint", Settings\Fingerprint)
        
      Case #Menu_About
        About::Open()
        
      Case #Menu_Exit
        Main\Quit = #True
        
    EndSelect
    
  EndProcedure
  
  Procedure Window_Event_Timer()
    Select EventTimer()
      Case 0
        Main()
        
      Case 1
        Window_StatusBar_Update()
        
    EndSelect
  EndProcedure
  
  Procedure Window_Event_CloseWindow()
    Main\Quit = 1
  EndProcedure
  
  Procedure Window_Open(Width, Height)
    
    Window\ID = OpenWindow(#PB_Any, 0, 0, Width, Height, #Software_Name + " V"+StrF(#Version*0.001,3), #PB_Window_SystemMenu | #PB_Window_SizeGadget | #PB_Window_TitleBar | #PB_Window_ScreenCentered | #PB_Window_MinimizeGadget | #PB_Window_MaximizeGadget)
    
    If Not Window\ID
      ProcedureReturn #False
    EndIf
    
    Window\Menu_ID = CreateImageMenu(#PB_Any, WindowID(Window\ID))
    If Not Window\Menu_ID
      MessageRequester(#Software_Name, "Couldn't create menu")
      CloseWindow(Window\ID)
      ProcedureReturn #False
    EndIf
    
    MenuTitle("File")
    MenuItem(#Menu_Exit, "Exit")
    
    MenuTitle("Templates")
    MenuItem(#Menu_Templates, "Edit", ImageID(Icon_images))
    
    MenuTitle("Canvas")
    MenuItem(#Menu_Canvas_Load, "Load viewport", ImageID(Icon_map))
    MenuItem(#Menu_Canvas_Reload, "Reload all", ImageID(Icon_map_go))
    
    MenuTitle("Settings")
    MenuItem(#Menu_Settings_Change_Fingerprint, "Change fingerprint", ImageID(Icon_key))
    
    MenuTitle("Help")
    ;MenuItem(#Menu_Dummy, "Hilfe")
    MenuItem(#Menu_About, "About", ImageID(Icon_information))
    
    ; #### Toolbar
    Window\ToolBar = CreateToolBar(#PB_Any, WindowID(Window\ID), #PB_ToolBar_Text)
    If Not Window\Menu_ID
      MessageRequester(#Software_Name, "Couldn't create toolbar")
      CloseWindow(Window\ID)
      ProcedureReturn #False
    EndIf
    ToolBarImageButton(#Menu_Templates, ImageID(Icon_images), #PB_ToolBar_Normal, "Templates") : ToolBarToolTip(Window\ToolBar, #Menu_Templates, "Manage templates")
    ToolBarImageButton(#Menu_Canvas_Load, ImageID(Icon_map), #PB_ToolBar_Normal, "Load viewport") : ToolBarToolTip(Window\ToolBar, #Menu_Canvas_Load, "Load all unloaded chunks inside the viewport")
    ToolBarImageButton(#Menu_Canvas_Reload, ImageID(Icon_map_go), #PB_ToolBar_Normal, "Reload all") : ToolBarToolTip(Window\ToolBar, #Menu_Canvas_Reload, "Reload all chunks")
    ToolBarImageButton(#Menu_Canvas_AutoReload, ImageID(Icon_time_go), #PB_ToolBar_Toggle, "Autoreload") : ToolBarToolTip(Window\ToolBar, #Menu_Canvas_AutoReload, "Reload all chunks every 60 minutes (Shouldn't be used anymore)")
    ToolBarImageButton(#Menu_Captcha_Requester, ImageID(Icon_bell), #PB_ToolBar_Toggle, "Captcha Requester") : ToolBarToolTip(Window\ToolBar, #Menu_Captcha_Requester, "Show a notification window, when a captcha has to be solved")
    ToolBarImageButton(#Menu_Settings_Change_Fingerprint, ImageID(Icon_key), #PB_ToolBar_Normal, "Fingerprint") : ToolBarToolTip(Window\ToolBar, #Menu_Settings_Change_Fingerprint, "Change fingerprint (Shouldn't be used anymore)")
    
    SetToolBarButtonState(Window\ToolBar, #Menu_Canvas_AutoReload, Settings\Canvas_AutoReload)
    SetToolBarButtonState(Window\ToolBar, #Menu_Captcha_Requester, Settings\Captcha_Requester)
    
    ; #### Shortcuts
    ;AddKeyboardShortcut(Window\ID, #PB_Shortcut_Control | #PB_Shortcut_F, #Menu_Search)
    ;AddKeyboardShortcut(Window\ID, #PB_Shortcut_F3, #Menu_Search_Continue)
    
    ; #### Statusbar
    Window\Statusbar = CreateStatusBar(#PB_Any, WindowID(Window\ID))
    If Not Window\Menu_ID
      MessageRequester(#Software_Name, "Couldn't create statusbar")
      CloseWindow(Window\ID)
      ProcedureReturn #False
    EndIf
    
    AddStatusBarField(100)
    AddStatusBarField(100)
    AddStatusBarField(100)
    AddStatusBarField(350)
    AddStatusBarField(150)
    AddStatusBarField(150)
    
    ; #### Timer
    AddWindowTimer(Window\ID, 0, 10)
    AddWindowTimer(Window\ID, 1, 100)
    
    ; #### Events
    BindEvent(#PB_Event_SizeWindow, @Window_Event_SizeWindow(), Window\ID)
    BindEvent(#PB_Event_Menu, @Window_Event_Menu(), Window\ID)
    BindEvent(#PB_Event_Timer, @Window_Event_Timer(), Window\ID)
    BindEvent(#PB_Event_CloseWindow, @Window_Event_CloseWindow(), Window\ID)
    
    ; #### Größe
    Window\Menu_Height = MenuHeight()
    Window\ToolBar_Height = ToolBarHeight(Window\ToolBar)
    Window\StatusBar_Height = StatusBarHeight(Window\StatusBar)
    
    ; #### Gadgets
    If UseGadgetList(WindowID(Window\ID))
      Window\Canvas\ID = CanvasGadget(#PB_Any, 0, Window\ToolBar_Height, Width, Height-Window\Menu_Height-Window\StatusBar_Height-Window\ToolBar_Height, #PB_Canvas_Keyboard)
      
      ; #### Gadget Events
      BindGadgetEvent(Window\Canvas\ID, @Canvas_Event())
    EndIf
    
    ; #### Redraw canvas
    Window\Canvas\Redraw = #True
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Image_Get(X.i, Y.i, Width.i, Height.i)
    Protected ImageID.i
    Protected R_X, R_Y
    
    Chunk_Collection_Download_Area(X, Y, Width, Height)
    
    ImageID = CreateImage(#PB_Any, Width, Height, 32, #PB_Image_Transparent)
    If ImageID And StartDrawing(ImageOutput(ImageID))
      DrawingMode(#PB_2DDrawing_AllChannels)
      
      ForEach Chunk_Collection()
        ForEach Chunk_Collection()\Chunk()
          ; #### Check if chunk is inside image area
          R_X = Chunk_Collection()\Chunk()\CX * #Chunk_Size - X
          R_Y = Chunk_Collection()\Chunk()\CY * #Chunk_Size - Y
          If Chunk_Collection()\Chunk()\Image And R_X + #Chunk_Size > 0 And R_X < Width And R_Y + #Chunk_Size > 0 And R_Y < Height
            DrawImage(ImageID(Chunk_Collection()\Chunk()\Image), R_X, R_Y)
          EndIf
        Next
      Next
      
      StopDrawing()
    EndIf
    
    ProcedureReturn ImageID
  EndProcedure
  
  Procedure Canvas_Redraw()
    Protected X_M.d, Y_M.d
    Protected Width, Height
    Protected Temp_Image_Scaled
    Protected *Chunk.Chunk
    
    If Not StartVectorDrawing(CanvasVectorOutput(Window\Canvas\ID))
      ProcedureReturn #False
    EndIf
    
    Width = VectorOutputWidth()
    Height = VectorOutputHeight()
    
    VectorSourceColor(RGBA(0, 0, 0, 255))
    FillVectorOutput()
    
    TranslateCoordinates(Int(Width / 2 + Settings\X), Int(Height / 2 + Settings\Y))
    ScaleCoordinates(Settings\Zoom, Settings\Zoom)
    
    ForEach Chunk_Collection()
      ForEach Chunk_Collection()\Chunk()
        *Chunk = Chunk_Collection()\Chunk()
        ; #### Check if chunk is inside 
        X_M = *Chunk\CX * #Chunk_Size * Settings\Zoom + Settings\X
        Y_M = *Chunk\CY * #Chunk_Size * Settings\Zoom + Settings\Y
        If X_M > Width / 2 Or Y_M > Height / 2 Or X_M < -Width / 2 - #Chunk_Size * Settings\Zoom Or Y_M < -Height / 2 - #Chunk_Size * Settings\Zoom
          ; #### Free scaled image
          If *Chunk\Image_Scaled
            FreeImage(*Chunk\Image_Scaled) : *Chunk\Image_Scaled = 0
            *Chunk\Image_Scale = 0
          EndIf
          *Chunk\Visible = #False
          Continue
        EndIf
        
        *Chunk\Visible = #True
        
        ; #### Resize image if needed
        If Settings\Zoom <> 1
          If *Chunk\Image_Scale <> Settings\Zoom
            If *Chunk\Image_Scaled
              FreeImage(*Chunk\Image_Scaled) : *Chunk\Image_Scaled = 0
              *Chunk\Image_Scale = 0
            EndIf
            If *Chunk\Image
              *Chunk\Image_Scaled = CopyImage(*Chunk\Image, #PB_Any)
              If *Chunk\Image_Scaled
                *Chunk\Image_Scale = Settings\Zoom
                If Settings\Zoom > 1
                  If Not ResizeImage(*Chunk\Image_Scaled, #Chunk_Size * Settings\Zoom, #Chunk_Size * Settings\Zoom, #PB_Image_Raw)
                    FreeImage(*Chunk\Image_Scaled) : *Chunk\Image_Scaled = 0
                    *Chunk\Image_Scale = 0
                  EndIf
                Else
                  If Not ResizeImage(*Chunk\Image_Scaled, #Chunk_Size * Settings\Zoom, #Chunk_Size * Settings\Zoom, #PB_Image_Smooth)
                    FreeImage(*Chunk\Image_Scaled) : *Chunk\Image_Scaled = 0
                    *Chunk\Image_Scale = 0
                  EndIf
                EndIf
              EndIf
            EndIf
          EndIf
        ElseIf *Chunk\Image_Scaled
          FreeImage(*Chunk\Image_Scaled) : *Chunk\Image_Scaled = 0
          *Chunk\Image_Scale = 0
        EndIf
        
        ; #### Draw chunk
        MovePathCursor(*Chunk\CX * #Chunk_Size, *Chunk\CY * #Chunk_Size)
        If *Chunk\Image_Scaled
          DrawVectorImage(ImageID(*Chunk\Image_Scaled), 255, #Chunk_Size, #Chunk_Size)
        ElseIf *Chunk\Image
          DrawVectorImage(ImageID(*Chunk\Image), 255, #Chunk_Size, #Chunk_Size)
        EndIf
      Next
    Next
    
    ; #### Draw templates
    ForEach Templates::Object()
      If (Templates::Object()\Settings\Active And Window\Canvas\Show_Templates) Or (Templates::Window\ID And GetGadgetState(Templates::Window\ListIcon) = ListIndex(Templates::Object()))
        X_M = Templates::Object()\Settings\X * Settings\Zoom + Settings\X
        Y_M = Templates::Object()\Settings\Y * Settings\Zoom + Settings\Y
        If X_M < Width / 2 And Y_M < Height / 2 And X_M > -Width / 2 - Templates::Object()\Width * Settings\Zoom And Y_M > -Height / 2 - Templates::Object()\Height * Settings\Zoom
          MovePathCursor(Templates::Object()\Settings\X, Templates::Object()\Settings\Y)
          If Templates::Object()\Template_ImageID
            Temp_Image_Scaled = CopyImage(Templates::Object()\Template_ImageID, #PB_Any)
            If Temp_Image_Scaled
              ResizeImage(Temp_Image_Scaled, Templates::Object()\Width * Settings\Zoom, Templates::Object()\Height * Settings\Zoom, #PB_Image_Raw)
              DrawVectorImage(ImageID(Temp_Image_Scaled), 255, Templates::Object()\Width, Templates::Object()\Height)
              FreeImage(Temp_Image_Scaled)
            EndIf
          EndIf
        EndIf
      EndIf
    Next
    
    StopVectorDrawing()
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Chunk_Get(CX.i, CY.i, Create=#False)
    Protected CCX.i = Quad_Divide_Floor(CX + #Chunk_Collection_Radius, #Chunk_Collection_Radius*2+1)
    Protected CCY.i = Quad_Divide_Floor(CY + #Chunk_Collection_Radius, #Chunk_Collection_Radius*2+1)
    
    Protected *Chunk_Collection.Chunk_Collection = Chunk_Collection_Get(CCX, CCY, Create)
    If Not *Chunk_Collection
      ProcedureReturn #Null
    EndIf
    
    ForEach *Chunk_Collection\Chunk()
      If *Chunk_Collection\Chunk()\CX = CX And *Chunk_Collection\Chunk()\CY = CY
        ProcedureReturn *Chunk_Collection\Chunk()
      EndIf
    Next
    
    ProcedureReturn #Null
  EndProcedure
  
  Procedure Chunk_Collection_Get(CCX.i, CCY.i, Create=#False)
    ForEach Chunk_Collection()
      If Chunk_Collection()\CCX = CCX And Chunk_Collection()\CCY = CCY
        ProcedureReturn Chunk_Collection()
      EndIf
    Next
    
    If Create
      ProcedureReturn _Chunk_Collection_Create(CCX, CCY)
    Else
      ProcedureReturn #Null
    EndIf
  EndProcedure
  
  Procedure Chunk_Collection_Finish_Handler(*Chunk_Collection.Chunk_Collection, *Memory)
    Protected ix, iy, jx, jy
    Protected CX.i, CY.i, CCX.i, CCY.i
    Protected Offset.i
    Protected Color, Color_Index
    Protected *DrawingBuffer.Ascii
    Protected *Chunk.Chunk
    
    CCX = *Chunk_Collection\CCX
    CCY = *Chunk_Collection\CCY
    
    Debug "Downloaded chunk collection: CCX:" + CCX + " CCY:" + CCY + " Radius:" + #Chunk_Collection_Radius
    
    If Not *Memory
      ProcedureReturn #False
    EndIf
    
    If MemorySize(*Memory) <> (#Chunk_Size * #Chunk_Size) * ((#Chunk_Collection_Radius * 2 + 1) * (#Chunk_Collection_Radius * 2 + 1)) / 2
      Debug "Downloaded chunk collection: Received memory size doesn't match!"
      ProcedureReturn #False
    EndIf
    
    ; #### Fill chunks in collection with image data
    For iy = 0 To #Chunk_Collection_Radius * 2
      For ix = 0 To #Chunk_Collection_Radius * 2
        CX = CCX * (#Chunk_Collection_Radius*2+1) + ix - #Chunk_Collection_Radius
        CY = CCY * (#Chunk_Collection_Radius*2+1) + iy - #Chunk_Collection_Radius
        *Chunk = Chunk_Get(CX, CY)
        If *Chunk
          If Not *Chunk\Image
            *Chunk\Image = CreateImage(#PB_Any, #Chunk_Size, #Chunk_Size, 32)
          EndIf
          
          If *Chunk\Image And StartDrawing(ImageOutput(*Chunk\Image))
            DrawingMode(#PB_2DDrawing_AllChannels)
            
            *DrawingBuffer = DrawingBuffer()
            For jy = #Chunk_Size-1 To 0 Step -1
              For jx = 0 To #Chunk_Size-1
                Offset = jx + jy * #Chunk_Size + (ix + iy * (#Chunk_Collection_Radius*2+1)) * #Chunk_Size * #Chunk_Size
                If Offset & 1
                  Color_Index = PeekA(*Memory + Offset / 2) & $0F
                Else
                  Color_Index = (PeekA(*Memory + Offset / 2) >> 4) & $0F
                EndIf
                *DrawingBuffer\a = Palette(Color_Index)\B : *DrawingBuffer + 1
                *DrawingBuffer\a = Palette(Color_Index)\G : *DrawingBuffer + 1
                *DrawingBuffer\a = Palette(Color_Index)\R : *DrawingBuffer + 1
                *DrawingBuffer + 1
              Next
            Next
            StopDrawing()
            Window\Canvas\Redraw = #True
          EndIf
        EndIf
      Next
    Next
    
    *Chunk_Collection\Downloaded = #True
    
    ForEach *Chunk_Collection\Chunk()
      Templates::Update_Chunk(*Chunk_Collection\Chunk())
    Next
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Chunk_Collection_Download_Handler(*Chunk_Collection.Chunk_Collection)
    Protected *Memory
    Static CacheFix_Counter
    Protected URL.s
    
    If *Chunk_Collection\Asynchronous_Download And *Chunk_Collection\Download_Timeout < ElapsedMilliseconds()
      AbortHTTP(*Chunk_Collection\Asynchronous_Download)
    EndIf
    
    If Not *Chunk_Collection\Asynchronous_Download And Not *Chunk_Collection\Downloaded
      
      URL = "http://pixelcanvas.io/api/bigchunk/"+Str(*Chunk_Collection\CCX*(#Chunk_Collection_Radius*2+1))+"."+Str(*Chunk_Collection\CCY*(#Chunk_Collection_Radius*2+1))+".bmp?CFC="+Str(CacheFix_Counter)
      
      *Chunk_Collection\Asynchronous_Download = ReceiveHTTPMemory(URL, #PB_HTTP_Asynchronous)
      CacheFix_Counter + 1
      
      *Chunk_Collection\Download_Timeout = ElapsedMilliseconds() + #Chunk_Collection_Download_Timeout
      
      Debug "Started chunk collection download: " + URL
      
    EndIf
    
    If Not *Chunk_Collection\Asynchronous_Download
      ProcedureReturn
    EndIf
    
    Select HTTPProgress(*Chunk_Collection\Asynchronous_Download)
      Case #PB_Http_Success
        *Memory = FinishHTTP(*Chunk_Collection\Asynchronous_Download) : *Chunk_Collection\Asynchronous_Download = 0
        If *Memory
          Chunk_Collection_Finish_Handler(*Chunk_Collection, *Memory)
          FreeMemory(*Memory)
        EndIf
        
      Case #PB_Http_Failed
        *Chunk_Collection\Asynchronous_Download = #False
        
      Case #PB_Http_Aborted
        *Chunk_Collection\Asynchronous_Download = #False
        
    EndSelect
  EndProcedure
  
  Procedure Chunk_Collection_Download_Area(X.i, Y.i, Width.i, Height.i)
    Protected ix, iy
    Protected CX_1.i, CY_1.i, CX_2.i, CY_2.i
    Protected CCX_1.i, CCY_1.i, CCX_2.i, CCY_2.i
    
    CX_1 = Quad_Divide_Floor(X, #Chunk_Size)
    CY_1 = Quad_Divide_Floor(Y, #Chunk_Size)
    CX_2 = Quad_Divide_Floor(X + Width, #Chunk_Size)
    CY_2 = Quad_Divide_Floor(Y + Height, #Chunk_Size)
    
    CCX_1 = Quad_Divide_Floor(CX_1 + #Chunk_Collection_Radius, #Chunk_Collection_Radius*2+1)
    CCY_1 = Quad_Divide_Floor(CY_1 + #Chunk_Collection_Radius, #Chunk_Collection_Radius*2+1)
    CCX_2 = Quad_Divide_Floor(CX_2 + #Chunk_Collection_Radius, #Chunk_Collection_Radius*2+1)
    CCY_2 = Quad_Divide_Floor(CY_2 + #Chunk_Collection_Radius, #Chunk_Collection_Radius*2+1)
    
    If CCX_1 > CCX_2 Or CCY_1 > CCY_2
      ProcedureReturn #False
    EndIf
    
    For iy = CCY_1 To CCY_2
      For ix = CCX_1 To CCX_2
        Debug "Chunk collection download area: CCX: " + ix + " CCY: " + iy
        Chunk_Collection_Get(ix, iy, #True)
      Next
    Next
    
    ProcedureReturn #True ; May happen, that not all chunk collections are downloaded
  EndProcedure
  
  Procedure Chunk_Collection_Delete(*Chunk_Collection.Chunk_Collection)
    ChangeCurrentElement(Chunk_Collection(), *Chunk_Collection)
    
    ForEach Chunk_Collection()\Chunk()
      If Chunk_Collection()\Chunk()\Image
        FreeImage(Chunk_Collection()\Chunk()\Image) : Chunk_Collection()\Chunk()\Image = 0
      EndIf
      If Chunk_Collection()\Chunk()\Image_Scaled
        FreeImage(Chunk_Collection()\Chunk()\Image_Scaled) : Chunk_Collection()\Chunk()\Image_Scaled = 0
        Chunk_Collection()\Chunk()\Image_Scale = 0
      EndIf
      
      DeleteElement(Chunk_Collection()\Chunk())
    Next
    
    DeleteElement(Chunk_Collection())
    
    Window\Canvas\Redraw = #True
  EndProcedure
  
  Procedure _Chunk_Collection_Create(CCX.i, CCY.i)
    Protected ix, iy
    
    AddElement(Chunk_Collection())
    
    Chunk_Collection()\CCX = CCX
    Chunk_Collection()\CCY = CCY
    
    For iy = -#Chunk_Collection_Radius To #Chunk_Collection_Radius
      For ix = -#Chunk_Collection_Radius To #Chunk_Collection_Radius
        AddElement(Chunk_Collection()\Chunk())
        
        Chunk_Collection()\Chunk()\CX = CCX * (#Chunk_Collection_Radius*2+1) + ix
        Chunk_Collection()\Chunk()\CY = CCY * (#Chunk_Collection_Radius*2+1) + iy
        
      Next
    Next
    
    Debug "Created chunk collection"
    
    ProcedureReturn Chunk_Collection()
  EndProcedure
  
  Procedure Update_Pixel(X.i, Y.i, Color.l)
    Protected CX.i = Quad_Divide_Floor(X, #Chunk_Size)
    Protected CY.i = Quad_Divide_Floor(Y, #Chunk_Size)
    Protected *Chunk.Chunk
    
    *Chunk = Chunk_Get(CX, CY)
    If *Chunk
      ;If Not Chunk()\Image
      ;  Chunk()\Image = CreateImage(#PB_Any, #Chunk_Size, #Chunk_Size, 32)
      ;EndIf
      If *Chunk\Image And StartDrawing(ImageOutput(*Chunk\Image))
        DrawingMode(#PB_2DDrawing_AllChannels)
        Plot(X - CX * #Chunk_Size, Y - CY * #Chunk_Size, Color)
        StopDrawing()
        *Chunk\Update = #True
      EndIf
    EndIf
    
    Templates::Update_Pixel(X, Y, Color)
    
    Main\Counter_PPS + 1
  EndProcedure
  
  ; !!!!!!!!!!!!!!!! Not used anymore !!!!!!!!!!!!!!!!
;   Procedure WebSocket_Send_TimestampRequest()
;     If Not WebSocket\Connection
;       ProcedureReturn #False
;     EndIf
;     
;     Protected Temp.l
;     PokeA(@Temp+0, $75)
;     PokeA(@Temp+1, Main\Timestamp_Counter >> 8)
;     PokeA(@Temp+2, Main\Timestamp_Counter)
;     Main\Timestamp_Counter + 1
;     
;     WebsocketClient::Frame_Send(WebSocket\Connection, @Temp, 3)
;     
;     Debug "Sent TimestampRequest: ID:" + Str(Main\Timestamp_Counter - 1)
;   EndProcedure
  
  ; !!!!!!!!!!!!!!!! Not used anymore !!!!!!!!!!!!!!!!
  Procedure WebSocket_Send_Input(X.w, Y.w, Color_Index.a, *Template.Templates::Object=#Null)
    If Not WebSocket\Connection
      ProcedureReturn #False
    EndIf
    
    ; #### Check if a pixel is already being placed
    ForEach Input_Blocking()
      If Input_Blocking()\X = X And Input_Blocking()\Y = Y
        ProcedureReturn #False
      EndIf
    Next
    
    Protected Temp.q
    
    PokeA(@Temp+0, $01)
    PokeA(@Temp+1, X >> 8)
    PokeA(@Temp+2, X)
    PokeA(@Temp+3, Y >> 8)
    PokeA(@Temp+4, Y)
    PokeA(@Temp+5, Color_Index)
    
    WebsocketClient::Frame_Send(WebSocket\Connection, @Temp, 6)
    
    LastElement(Input_Blocking())
    AddElement(Input_Blocking())
    Input_Blocking()\X = X
    Input_Blocking()\Y = Y
    Input_Blocking()\Color_Index = Color_Index
    Input_Blocking()\Timestamp = ElapsedMilliseconds()
    
    Debug "Sent Input: X:" + Str(X) + " Y:" + Str(Y) + " Color_Index:" + Str(Color_Index)
    StatusBarText(Window\StatusBar, 3, "Last: X:" + X + " Y:" + Y + " Color:" + Color_Index)
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure WebSocket_Client_Main()
    Protected i
    Static Time_Ping
    Protected NetworkEvent
    Protected FrameType.u
    Protected *Data
    Protected CX.w, CY.w, X, Y, Other, Color.l, Color_Index.a
    Protected Timestamp.d
    Protected Timestamp_ID.u
    Protected *Template.Templates::Object
    Protected *Temp_Memory, JSON_Temp
    
    If WebSocket\Connection
      
      If Time_Ping < ElapsedMilliseconds()
        Time_Ping = ElapsedMilliseconds() + 60000
        WebsocketClient::Frame_Send(WebSocket\Connection, #Null, 0, WebSocketClient::#Opcode_Ping)
      EndIf
      
      Repeat
        
        NetworkEvent = NetworkClientEvent(WebSocket\Connection)
        
        Select NetworkEvent
          Case #PB_NetworkEvent_Data
            *Data = WebsocketClient::Frame_Receive(WebSocket\Connection, @FrameType)
            If *Data
              Select FrameType
                Case WebsocketClient::#Opcode_Binary
                  Select PeekA(*Data)
                    Case $75 ; Current timestamp
                      ; !!!!!!!!!!!!!!!! Not used anymore !!!!!!!!!!!!!!!!
;                       Timestamp_ID = PeekA(*Data+1) << 8 + PeekA(*Data+2)
;                       For i = 0 To 7
;                         PokeA(@Timestamp + i, PeekA(*Data + 3 + 7 - i))
;                       Next
;                       ; #### Check if the timestamp is the answer to the latest request
;                       If Timestamp_ID = Main\Timestamp_Counter - 1
;                         Main\Timestamp_Offset = Timestamp - Main\Timestamp_TimeRequest
;                       EndIf
;                       Debug "Server Timestamp: ID:" + Str(Timestamp_ID) + " Time: " + StrD(Timestamp)
                      
                    Case $57 ; Timestamp for next pixel
                      ; !!!!!!!!!!!!!!!! Not used anymore !!!!!!!!!!!!!!!!
;                       For i = 0 To 7
;                         PokeA(@Timestamp + i, PeekA(*Data + 1 + 7 - i))
;                       Next
;                       Main\Timestamp_Next_Pixel = Timestamp
;                       Debug "Timestamp Next Pixel: " + StrD(Timestamp)
;                       ; #### Update counter value
;                       If LastElement(Input_Check())
;                         *Template = Input_Check()\Template
;                         If *Template
;                           *Template\Settings\Counter + 1
;                           If Timestamp > Get_Timestamp() And Timestamp - Get_Timestamp() < 3600000
;                             *Template\Settings\Total_Time + (Timestamp - Get_Timestamp())
;                           EndIf
;                         EndIf
;                       EndIf
                      
                    Case $C1 ; Pixel update
                      CX = PeekA(*Data+1) << 8 + PeekA(*Data+2)
                      CY = PeekA(*Data+3) << 8 + PeekA(*Data+4)
                      Other = PeekA(*Data+5) << 8 + PeekA(*Data+6)
                      X = CX * #Chunk_Size + (Other >> 4) & $3F
                      Y = CY * #Chunk_Size + (Other >> 10) & $3F
                      Color_Index = PeekA(*Data+6) & $0F
                      Color = Palette(Color_Index)\Color
                      Update_Pixel(X, Y, Color)
                      ;Debug "Update Pixel: X:" + X + " Y:" + Y + " Color_Index:" + Str(Color_Index)
                      ; #### Pixel input check
                      ForEach Input_Blocking()
                        If Input_Blocking()\X = X And Input_Blocking()\Y = Y
                          DeleteElement(Input_Blocking())
                        EndIf
                      Next
                      
                  EndSelect
              EndSelect
              
              FreeMemory(*Data)
            Else
              Select FrameType
                Case WebSocketClient::#Opcode_Pong
                  Debug "Websocket Pong"
                Case WebSocketClient::#Opcode_Ping
                  Debug "Websocket Ping"
              EndSelect
            EndIf
            
          Case #PB_NetworkEvent_Disconnect
            WebSocket\Connection = 0
            Break
            
          Case #PB_NetworkEvent_None
            Break
            
        EndSelect
        
      ForEver
    ElseIf WebSocket\Reconnect_Timer < ElapsedMilliseconds()
      WebSocket\Reconnect_Timer = ElapsedMilliseconds() + #WebSocket_Reconnect_Time
      *Temp_Memory = ReceiveHTTPMemory("http://pixelcanvas.io/api/ws")
      If *Temp_Memory
        JSON_Temp = CatchJSON(#PB_Any, *Temp_Memory, MemorySize(*Temp_Memory))
        If JSON_Temp
          WebSocket\URL = GetJSONString(GetJSONMember(JSONValue(JSON_Temp), "url")) + "/?fingerprint=" + Settings\Fingerprint
          Debug "Catched websocket url: " + WebSocket\URL
          
          WebSocket\Connection = WebsocketClient::OpenWebsocketConnection(WebSocket\URL)
          ; #### Free all chunks, so they will be redownloaded correctly
          If WebSocket\Connection
            Debug "Connected to " + WebSocket\URL
            ; #### Reload all old chunks
            ForEach Chunk_Collection()
              Chunk_Collection()\Downloaded = #False
            Next
          Else
            Debug "Failed to connect to " + WebSocket\URL
          EndIf
          
          FreeJSON(JSON_Temp)
        EndIf
        
        FreeMemory(*Temp_Memory)
      EndIf
    EndIf
  EndProcedure
  
  #HTTP_ADDREQ_FLAG_ADD = $20000000
  #HTTP_ADDREQ_FLAG_ADD_IF_NEW = $10000000
  #HTTP_ADDREQ_FLAG_COALESCE_WITH_COMMA = $40000000
  #HTTP_ADDREQ_FLAG_COALESCE_WITH_SEMICOLON = $01000000
  #HTTP_ADDREQ_FLAG_COALESCE = #HTTP_ADDREQ_FLAG_COALESCE_WITH_COMMA
  #HTTP_ADDREQ_FLAG_REPLACE = $80000000
  #INTERNET_COOKIE_EVALUATE_P3P = $80
  #INTERNET_COOKIE_THIRD_PARTY = $10
  #INTERNET_FLAG_ASYNC = $10000000
  #INTERNET_FLAG_CACHE_ASYNC = $00000080
  #INTERNET_FLAG_CACHE_IF_NET_FAIL = $00010000
  #INTERNET_FLAG_DONT_CACHE = $04000000
  #INTERNET_FLAG_EXISTING_CONNECT = $20000000
  #INTERNET_FLAG_FORMS_SUBMIT = $00000040
  #INTERNET_FLAG_FROM_CACHE = $01000000
  #INTERNET_FLAG_FWD_BACK = $00000020
  #INTERNET_FLAG_HYPERLINK = $00000400
  #INTERNET_FLAG_IGNORE_CERT_CN_INVALID = $00001000
  #INTERNET_FLAG_IGNORE_CERT_DATE_INVALID = $00002000
  #INTERNET_FLAG_IGNORE_REDIRECT_TO_HTTP = $00008000
  #INTERNET_FLAG_IGNORE_REDIRECT_TO_HTTPS = $00004000
  #INTERNET_FLAG_KEEP_CONNECTION = $00400000
  #INTERNET_FLAG_MAKE_PERSISTENT = $02000000
  #INTERNET_FLAG_MUST_CACHE_REQUEST = $00000010
  #INTERNET_FLAG_NEED_FILE = $00000010
  #INTERNET_FLAG_NO_AUTH = $00040000
  #INTERNET_FLAG_NO_AUTO_REDIRECT = $00200000
  #INTERNET_FLAG_NO_CACHE_WRITE = $04000000
  #INTERNET_FLAG_NO_COOKIES = $00080000
  #INTERNET_FLAG_NO_UI = $00000200
  #INTERNET_FLAG_OFFLINE = $01000000
  #INTERNET_FLAG_PASSIVE = $08000000
  #INTERNET_FLAG_PRAGMA_NOCACHE = $00000100
  #INTERNET_FLAG_RAW_DATA = $40000000
  #INTERNET_FLAG_READ_PREFETCH = $00100000
  #INTERNET_FLAG_RELOAD = $80000000
  #INTERNET_FLAG_RESTRICTED_ZONE = $00020000
  #INTERNET_FLAG_RESYNCHRONIZE = $00000800
  #INTERNET_FLAG_SECURE = $00800000
  #INTERNET_FLAG_TRANSFER_ASCII = $00000001
  #INTERNET_FLAG_TRANSFER_BINARY = $00000002
  #INTERNET_NO_CALLBACK = $00000000
  #INTERNET_OPTION_SUPPRESS_SERVER_AUTH = 104
  #WININET_API_FLAG_ASYNC = $00000001
  #WININET_API_FLAG_SYNC = $00000004
  #WININET_API_FLAG_USE_CONTEXT = $00000008
  Procedure.s HTTP_Post(URL.s, Post_Data.s, Timeout.l)
    Protected Dim rgpszAcceptTypes.s(1)
    
    Protected Protocol.s = GetURLPart(URL, #PB_URL_Protocol)
    Protected Servername.s = GetURLPart(URL, #PB_URL_Site)
    Protected Port.l = Val(GetURLPart(URL, #PB_URL_Port))
    If Port.l = 0 : Port.l = 80 : EndIf
    Protected Path.s = GetURLPart(URL, #PB_URL_Path)
    If Path.s = "" : Path.s = "/" : EndIf
    Protected Parameters.s = GetURLPart(URL, #PB_URL_Parameters)
    Protected Address.s = Path
    
    Protected *Temp_Data = UTF8(Post_Data)
    Debug "Request: " + Post_Data
    
    Protected User_agent.s = "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:53.0) Gecko/20100101 Firefox/53.0"
    Protected Open_handle = InternetOpen_(User_agent, 1, "", "", 0)
    InternetSetOption_(Open_handle, 2, Timeout, 4)
    Protected Flags.l = #INTERNET_FLAG_KEEP_CONNECTION
    Protected Connect_handle = InternetConnect_(Open_handle, Servername, Port, #Null$, #Null$, 3, 0, 0)
    
    rgpszAcceptTypes(0) = "*/*"
    rgpszAcceptTypes(1) = #Null$
    
    Protected Request_handle = HttpOpenRequest_(Connect_handle, "POST", Path, #Null$, "http://pixelcanvas.io/@338,1970", @rgpszAcceptTypes(), Flags, 0)
    
    Protected Temp_Header.s
    ;Temp_Header.s = "Host: " + Servername + #CRLF$
    ;HttpAddRequestHeaders_(Request_handle, @Temp_Header, Len(Temp_Header), #HTTP_ADDREQ_FLAG_ADD | #HTTP_ADDREQ_FLAG_REPLACE)
    Temp_Header = "Accept-Language: de,en-US;q=0.7,en;q=0.3" + #CRLF$
    HttpAddRequestHeaders_(Request_handle, @Temp_Header, Len(Temp_Header), #HTTP_ADDREQ_FLAG_ADD | #HTTP_ADDREQ_FLAG_REPLACE)
    Temp_Header = "Content-Type: application/json" + #CRLF$
    HttpAddRequestHeaders_(Request_handle, @Temp_Header, Len(Temp_Header), #HTTP_ADDREQ_FLAG_ADD | #HTTP_ADDREQ_FLAG_REPLACE)
    ;Temp_Header = "Accept-Encoding: gzip, deflate" + #CRLF$
    ;HttpAddRequestHeaders_(Request_handle, @Temp_Header, Len(Temp_Header), #HTTP_ADDREQ_FLAG_ADD | #HTTP_ADDREQ_FLAG_REPLACE)
    Temp_Header = "Origin: http://pixelcanvas.io" + #CRLF$
    HttpAddRequestHeaders_(Request_handle, @Temp_Header, Len(Temp_Header), #HTTP_ADDREQ_FLAG_ADD | #HTTP_ADDREQ_FLAG_REPLACE)
    ;Temp_Header = "DNT: 1" + #CRLF$
    ;HttpAddRequestHeaders_(Request_handle, @Temp_Header, Len(Temp_Header), #HTTP_ADDREQ_FLAG_ADD | #HTTP_ADDREQ_FLAG_REPLACE)
    
    Protected Send_handle = HttpSendRequest_(Request_handle, #Null$, 0, *Temp_Data, MemorySize(*Temp_Data)-1)
    
    Protected Response.s = ""
    Protected Result.l, Bytes_Read.i
    Protected Buffer_Size = 1024
    Protected *Temp_Buffer = AllocateMemory(Buffer_Size)
    Repeat
      Result = InternetReadFile_(Request_handle, *Temp_Buffer, Buffer_Size, @Bytes_Read)
      If Result
        Response + PeekS(*Temp_Buffer, Bytes_Read, #PB_UTF8 | #PB_ByteLength)
      Else
        Debug "InternetReadFile_(...) results in #False"
      EndIf
    Until Bytes_Read = 0 Or Result = #False
    FreeMemory(*Temp_Buffer)
    
    InternetCloseHandle_(Open_handle)
    InternetCloseHandle_(Connect_handle)
    InternetCloseHandle_(Request_handle)
    InternetCloseHandle_(Send_handle)
    
    FreeMemory(*Temp_Data)
    
    Debug "Response: " + Response
    
    ProcedureReturn Response
  EndProcedure
  
  Procedure HTTP_Post_Input(X, Y, Color_Index.a, *Template.Templates::Object=#Null, Fingerprint.s="")
    Protected Error.s
    Protected Result
    
    ; #### Check if the input is blocked
    ForEach Input_Blocking()
      If Input_Blocking()\X = X And Input_Blocking()\Y = Y And Input_Blocking()\Color_Index = Color_Index
        ProcedureReturn #Input_Result_Local_Error
      EndIf
    Next
    
    ; #### Read cookie to get duck
;     Protected Cookie_Size.i, Duck.s
;     InternetGetCookie_("http://pixelcanvas.io", #Null, #Null, @Cookie_Size)
;     If Cookie_Size
;       Protected *Cookie_Buffer = AllocateMemory(Cookie_Size)
;       InternetGetCookie_("http://pixelcanvas.io", #Null, *Cookie_Buffer, @Cookie_Size)
;       If ExamineRegularExpression(RegEx_Duck, PeekS(*Cookie_Buffer, Cookie_Size))
;         If NextRegularExpressionMatch(RegEx_Duck)
;           Duck = RegularExpressionNamedGroup(RegEx_Duck, "Duck")
;         EndIf
;       EndIf
;       FreeMemory(*Cookie_Buffer)
;     Else
;       ProcedureReturn #Input_Result_Global_Error
;     EndIf
    
    Protected JSON = CreateJSON(#PB_Any)
    Protected JSON_Object = SetJSONObject(JSONValue(JSON))
    SetJSONInteger(AddJSONMember(JSON_Object, "x"), X)
    SetJSONInteger(AddJSONMember(JSON_Object, "y"), Y)
    ;SetJSONInteger(AddJSONMember(JSON_Object, Duck), X + Y + 24)
    SetJSONInteger(AddJSONMember(JSON_Object, "a"), X + Y + 8)
    SetJSONInteger(AddJSONMember(JSON_Object, "color"), Color_Index)
    SetJSONString(AddJSONMember(JSON_Object, "fingerprint"), Fingerprint)
    SetJSONNull(AddJSONMember(JSON_Object, "token"))
    
    Protected Response.s = HTTP_Post("http://pixelcanvas.io/api/pixel", ComposeJSON(JSON), 1000)
    FreeJSON(JSON)
    
    JSON = ParseJSON(#PB_Any, Response)
    If JSON
      If GetJSONMember(JSONValue(JSON), "success")
        Protected Success = GetJSONBoolean(GetJSONMember(JSONValue(JSON), "success"))
      EndIf
      If GetJSONMember(JSONValue(JSON), "waitSeconds") And JSONType(GetJSONMember(JSONValue(JSON), "waitSeconds")) = #PB_JSON_Number
        Userdata\Timestamp_Next_Pixel = Get_Timestamp() + GetJSONInteger(GetJSONMember(JSONValue(JSON), "waitSeconds")) * 1000
      EndIf
      If GetJSONMember(JSONValue(JSON), "errors")
        If GetJSONElement(GetJSONMember(JSONValue(JSON), "errors"), 0)
          If GetJSONMember(GetJSONElement(GetJSONMember(JSONValue(JSON), "errors"), 0), "msg")
            Error = GetJSONString(GetJSONMember(GetJSONElement(GetJSONMember(JSONValue(JSON), "errors"), 0), "msg"))
            Select Error
              Case "You are using a proxy!!!11!"
                Userdata\Logged_In = #False ; Cause a re-login
              Case "You must provide a token"
                Userdata\Timestamp_Next_Pixel = Get_Timestamp() + 1000 * 60
                If Settings\Captcha_Requester
                  Captcha_Requester::Open()
                EndIf
              Case "You must wait"
                Captcha_Requester::Close()
            EndSelect
          EndIf
        EndIf
      EndIf
      FreeJSON(JSON)
    EndIf
    
    If Not Success
      LastElement(Input_Blocking())
      AddElement(Input_Blocking())
      Input_Blocking()\X = X
      Input_Blocking()\Y = Y
      Input_Blocking()\Color_Index = Color_Index
      Input_Blocking()\Timestamp = ElapsedMilliseconds()
      StatusBarText(Window\StatusBar, 3, "Error: " + Error)
      
      Result = #Input_Result_Global_Error
    Else
      StatusBarText(Window\StatusBar, 3, "Last: X:" + X + " Y:" + Y + " Color:" + Color_Index + " '"+*Template\Settings\Filename+"'")
      Update_Pixel(X, Y, Palette(Color_Index)\Color)
      
      *Template\Settings\Counter + 1
      If Userdata\Timestamp_Next_Pixel > Get_Timestamp() And Userdata\Timestamp_Next_Pixel - Get_Timestamp() < 3600000
        *Template\Settings\Total_Time + (Userdata\Timestamp_Next_Pixel - Get_Timestamp())
      EndIf
      
      Result = #Input_Result_Success
      Captcha_Requester::Close()
    EndIf
    
    ProcedureReturn Result
  EndProcedure
  
  ; !!!!!!!!!!!!!!!! Not used anymore !!!!!!!!!!!!!!!!
;   Procedure HTTP_Post_Timesync(Fingerprint.s="")
;     Protected Start = ElapsedMilliseconds()
;     
;     Protected JSON = CreateJSON(#PB_Any)
;     Protected JSON_Object = SetJSONObject(JSONValue(JSON))
;     SetJSONString(AddJSONMember(JSON_Object, "jsonrpc"), "2.0")
;     SetJSONInteger(AddJSONMember(JSON_Object, "id"), Main\Timestamp_Counter) : Main\Timestamp_Counter + 1
;     SetJSONString(AddJSONMember(JSON_Object, "method"), "timesync")
;     
;     Protected Response.s = HTTP_Post("http://pixelcanvas.io/api/timesync", ComposeJSON(JSON), 1000)
;     FreeJSON(JSON)
;     
;     Main\Timestamp_TimeRequest = ElapsedMilliseconds()
;     
;     JSON = ParseJSON(#PB_Any, Response)
;     If JSON
;       If GetJSONMember(JSONValue(JSON), "id")
;         Protected ID = GetJSONInteger(GetJSONMember(JSONValue(JSON), "id"))
;       Else
;         Debug "Timesync failed, no id returned"
;       EndIf
;       If GetJSONMember(JSONValue(JSON), "result")
;         Protected Timestamp = GetJSONInteger(GetJSONMember(JSONValue(JSON), "result"))
;         Main\Timestamp_Offset = Timestamp - Main\Timestamp_TimeRequest
;       Else
;         Debug "Timesync failed, no result returned"
;       EndIf
;       FreeJSON(JSON)
;     EndIf
;     
;   EndProcedure
  
  Procedure HTTP_Post_Me(Fingerprint.s="")
    Protected Time_Request = ElapsedMilliseconds()
    
    Protected JSON = CreateJSON(#PB_Any)
    Protected JSON_Object = SetJSONObject(JSONValue(JSON))
    SetJSONString(AddJSONMember(JSON_Object, "fingerprint"), Fingerprint)
    
    Protected Response.s = HTTP_Post("http://pixelcanvas.io/api/me", ComposeJSON(JSON), 1000)
    FreeJSON(JSON)
    
    Protected Time_Response = ElapsedMilliseconds()
    
    JSON = ParseJSON(#PB_Any, Response)
    If JSON
      If GetJSONMember(JSONValue(JSON), "id")
        Userdata\ID = GetJSONString(GetJSONMember(JSONValue(JSON), "id"))
      EndIf
      If GetJSONMember(JSONValue(JSON), "name")
        Userdata\Name = GetJSONString(GetJSONMember(JSONValue(JSON), "name"))
      EndIf
      ;If GetJSONMember(JSONValue(JSON), "region")
      ;  Userdata\Region = GetJSONString(GetJSONMember(JSONValue(JSON), "region"))
      ;EndIf
      If GetJSONMember(JSONValue(JSON), "center")
        Userdata\Center_X = GetJSONInteger(GetJSONElement(GetJSONMember(JSONValue(JSON), "center"), 0))
        Userdata\Center_Y = GetJSONInteger(GetJSONElement(GetJSONMember(JSONValue(JSON), "center"), 1))
      EndIf
      If GetJSONMember(JSONValue(JSON), "waitSeconds") And JSONType(GetJSONMember(JSONValue(JSON), "waitSeconds")) = #PB_JSON_Number
        Userdata\Timestamp_Next_Pixel = Get_Timestamp() + GetJSONDouble(GetJSONMember(JSONValue(JSON), "waitSeconds")) * 1000
      EndIf
      ;If GetJSONMember(JSONValue(JSON), "serverTime")
      ;  Protected Timestamp = GetJSONInteger(GetJSONMember(JSONValue(JSON), "serverTime"))
      ;  Main\Timestamp_Offset = Timestamp - Time_Response;(Time_Request + Time_Response) / 2
      ;EndIf
      
      Userdata\Logged_In = #True
      
      FreeJSON(JSON)
    EndIf
    
  EndProcedure
  
  Procedure Main()
    Static AutoClear_Timer.q
    Protected CX.i, CY.i, CCX.i, CCY.i
    Protected *Chunk.Chunk
    
    WebSocket_Client_Main()
    
    ; #### Blinking of templates
    If Window\Canvas\Show_Templates_Blink_Timer < ElapsedMilliseconds()
      Window\Canvas\Show_Templates_Blink_Timer = ElapsedMilliseconds() + 500
      If Window\Canvas\Show_Templates
        Window\Canvas\Show_Templates = #False
      Else
        Window\Canvas\Show_Templates = #True
      EndIf
      Window\Canvas\Redraw = #True
    EndIf
    
    ; #### AutoClear
    If Settings\Canvas_AutoReload And AutoClear_Timer < ElapsedMilliseconds()
      AutoClear_Timer = ElapsedMilliseconds() + 60000 * 30
      ForEach Chunk_Collection()
        Chunk_Collection()\Downloaded = #False
      Next
    EndIf
    
    ; #### Login
    If Not Userdata\Logged_In
      HTTP_Post_Me(Settings\Fingerprint)
    EndIf
    
    ; #### Timesync request
    ;If Main\Timestamp_Next_TimeRequest < Date()
    ;  Main\Timestamp_Next_TimeRequest = Date() + 30
    ;  
    ;  HTTP_Post_Timesync(Settings\Fingerprint)
    ;  
    ;EndIf
    
    ; #### Download chunk collections
    ForEach Chunk_Collection()
      Chunk_Collection_Download_Handler(Chunk_Collection())
    Next
    
    ; #### Do stuff with chunks
    ForEach Chunk_Collection()
      ForEach Chunk_Collection()\Chunk()
        *Chunk = Chunk_Collection()\Chunk()
        If *Chunk\Update
          *Chunk\Update = #False
          
          ; #### Free scaled image as it isn't up to date anymore
          If *Chunk\Image_Scaled
            FreeImage(*Chunk\Image_Scaled) : *Chunk\Image_Scaled = 0
            *Chunk\Image_Scale = 0
          EndIf
          
          If *Chunk\Visible
            Window\Canvas\Redraw = #True
          EndIf
        EndIf
        
      Next
    Next
    
    ; #### Delete input check entries after some time
    ForEach Input_Blocking()
      If Input_Blocking()\Timestamp + 1000 * 60 * 10 < ElapsedMilliseconds()
;         CX = Quad_Divide_Floor(Input_Check()\X, #Chunk_Size)
;         CY = Quad_Divide_Floor(Input_Check()\Y, #Chunk_Size)
;         CCX = Quad_Divide_Floor(CX + #Chunk_Collection_Radius, #Chunk_Collection_Radius*2+1)
;         CCY = Quad_Divide_Floor(CY + #Chunk_Collection_Radius, #Chunk_Collection_Radius*2+1)
;         If Chunk_Collection_Get(CCX, CCY)
;           Chunk_Collection_Delete(Chunk_Collection())
;         EndIf
        DeleteElement(Input_Blocking())
      EndIf
    Next
    
    If Main\Timer_PPS + 5000 < ElapsedMilliseconds()
      
      Main\PPS = Main\Counter_PPS / (ElapsedMilliseconds() - Main\Timer_PPS) * 1000
      Main\Counter_PPS = 0
      
      Main\Timer_PPS = ElapsedMilliseconds()
      
      StatusBarText(Window\StatusBar, 5, StrD(Main\PPS, 1) + " Pixels/s", #PB_StatusBar_Right)
    EndIf
    
    If Window\Canvas\Redraw And GetWindowState(Window\ID) <> #PB_Window_Minimize
      Window\Canvas\Redraw = #False
      Canvas_Redraw()
    EndIf
    
    Templates::Main()
    Captcha_Requester::Main()
    About::Main()
    
  EndProcedure
  
  ; ################################################### Initialisation ##############################################
  If FileSize("Portable/") = -2
    Main\Path_AppData = "Portable/"
  Else
    Main\Path_AppData = Helper::SHGetFolderPath(#CSIDL_APPDATA) + "/D3/Pixelcanvas Client/"
    MakeSureDirectoryPathExists(Main\Path_AppData)
  EndIf
  
  Settings_Load(Main\Path_AppData + #Filename_Settings)
  Templates::Settings_Load(Main\Path_AppData + #Filename_Templates)
  
  If Not Window_Open(1000, 600)
    End
  EndIf
  
  ; ################################################### Main ########################################################
  Repeat
    
    While WaitWindowEvent(1)
    Wend
    
    Main()
    
  Until Main\Quit
  
  ; ################################################### End #########################################################
  Templates::Settings_Save(Main\Path_AppData + #Filename_Templates)
  Settings_Save(Main\Path_AppData + #Filename_Settings)
  
  ; ################################################### Data Sections ###############################################
  DataSection
    Icon_images:        : IncludeBinary "Data/Icons/images.png"
    Icon_map:           : IncludeBinary "Data/Icons/map.png"
    Icon_map_go:        : IncludeBinary "Data/Icons/map_go.png"
    Icon_time_go:       : IncludeBinary "Data/Icons/time_go.png"
    Icon_key:           : IncludeBinary "Data/Icons/key.png"
    Icon_information:   : IncludeBinary "Data/Icons/information.png"
    Icon_bell:          : IncludeBinary "Data/Icons/bell.png"
  EndDataSection
  
EndModule
; IDE Options = PureBasic 5.60 (Windows - x64)
; CursorPosition = 1237
; FirstLine = 1205
; Folding = -----
; EnableThread
; EnableXP
; EnableUser
; Executable = Pixelcanvas Client.exe
; DisableDebugger
; EnablePurifier = 1,1,1,1
; EnableCompileCount = 512
; EnableBuildCount = 70
; EnableExeConstant