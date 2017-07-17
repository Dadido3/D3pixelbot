; ##################################################### License / Copyright #########################################
; 
; ##################################################### Dokumentation / Kommentare ##################################
; 
; 
; 
; 
; 
; 
; 
; ##################################################### Includes ####################################################

; ###################################################################################################################
; ##################################################### Public ######################################################
; ###################################################################################################################

DeclareModule Captcha_Requester
  EnableExplicit
  ; ################################################### Constants ###################################################
  
  ; ################################################### Functions ###################################################
  Declare   Open()
  Declare   Close()
  Declare   Main()
  
EndDeclareModule

; ###################################################################################################################
; ##################################################### Private #####################################################
; ###################################################################################################################

Module Captcha_Requester
  ; ##################################################### Includes ####################################################
  
  ; ##################################################### Prototypes ##################################################
  
  ; ##################################################### Structures ##################################################
  
  ; ##################################################### Constants ###################################################
  
  ; ##################################################### Structures ##################################################
  Structure Main
    
  EndStructure
  
  Structure Window
    ID.i
    Exit.i
    
    ; #### Gadgets
    Text.i
  EndStructure
  
  Structure FLASHWINFO_B Align #PB_Structure_AlignC
    cbSize.l
    hwnd.i
    dwFlags.l
    uCount.l
    dwTimeout.l
  EndStructure
  
  ; ##################################################### Variables ###################################################
  Global Main.Main
  Global Window.Window
  
  ; ##################################################### Icons ... ###################################################
  
  ; ##################################################### Init ########################################################
  ;Global Font = LoadFont(#PB_Any, "Courier New", 20)
  
  ; ##################################################### Procedures ##################################################
  Procedure Window_Flash(hwnd.i)
    Protected Info.FLASHWINFO
    Info\cbSize = SizeOf(FLASHWINFO_B)
    Info\hwnd   = hwnd
    Info\dwFlags= #FLASHW_CAPTION | #FLASHW_TRAY | #FLASHW_TIMER
    Info\uCount = 10
    Info\dwTimeout = 300
    
    FlashWindowEx_(Info)
  EndProcedure
  
  Procedure Event_SizeWindow()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
  EndProcedure
  
  Procedure Event_ActivateWindow()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
  EndProcedure
  
  Procedure Event_Menu()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    Protected Event_Menu = EventMenu()
    
  EndProcedure
  
  Procedure Event_CloseWindow()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    Window\Exit = #True
  EndProcedure
  
  Procedure Open()
    Protected Width, Height
    
    Beep_(1000, 50)
    Beep_(2000, 50)
    Beep_(4000, 50)
    
    If Window\ID
      ;SetActiveWindow(Window\ID)
      Window_Flash(WindowID(Window\ID))
      ProcedureReturn #True
    EndIf
    
    Width = 200
    Height = 100
    
    Window\ID = OpenWindow(#PB_Any, 0, 0, Width, Height, "Captcha", #PB_Window_SystemMenu | #PB_Window_ScreenCentered)
    If Window\ID
      
      Window\Text = TextGadget(#PB_Any, 10, 10, Width - 20, Height -20, "The captcha has to be solved manually." + #CRLF$ + "Go to PixelCanvas.io, place a pixel and solve the captcha(s)." + #CRLF$ + "This window will disappear then eventually.")
      
      BindEvent(#PB_Event_SizeWindow, @Event_SizeWindow(), Window\ID)
      BindEvent(#PB_Event_Menu, @Event_Menu(), Window\ID)
      BindEvent(#PB_Event_CloseWindow, @Event_CloseWindow(), Window\ID)
      
      Window_Flash(WindowID(Window\ID))
      
      ProcedureReturn #True
    EndIf
    
    ProcedureReturn #False
  EndProcedure
  
  Procedure Close()
    If Window\ID
      
      UnbindEvent(#PB_Event_SizeWindow, @Event_SizeWindow(), Window\ID)
      UnbindEvent(#PB_Event_Menu, @Event_Menu(), Window\ID)
      UnbindEvent(#PB_Event_CloseWindow, @Event_CloseWindow(), Window\ID)
      
      CloseWindow(Window\ID)
      Window\ID = 0
    EndIf
  EndProcedure
  
  Procedure Main()
    If Window\Exit
      Window\Exit = #False
      Close()
    EndIf
    
  EndProcedure
  
  ; ##################################################### Initialisation ##############################################
  
  ; ##################################################### Data Sections ###############################################
  
EndModule

; IDE Options = PureBasic 5.60 beta 6 (Windows - x64)
; CursorPosition = 117
; FirstLine = 100
; Folding = --
; EnableXP