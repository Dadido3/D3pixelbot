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

DeclareModule Goto_Requester
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

Module Goto_Requester
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
    Text.i [10]
    Spin.i [10]
    Button.i
  EndStructure
  
  ; ##################################################### Variables ###################################################
  Global Main.Main
  Global Window.Window
  
  ; ##################################################### Icons ... ###################################################
  
  ; ##################################################### Init ########################################################
  ;Global Font = LoadFont(#PB_Any, "Courier New", 20)
  
  ; ##################################################### Procedures ##################################################
  Procedure Button_Event()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    Select Event_Type
      Case #PB_EventType_LeftClick
        Main::Canvas_Goto(-GetGadgetState(Window\Spin[0]), -GetGadgetState(Window\Spin[1]))
        
    EndSelect
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
    
    If Window\ID
      SetActiveWindow(Window\ID)
      ProcedureReturn #True
    EndIf
    
    Width = 150
    Height = 100
    
    Window\ID = OpenWindow(#PB_Any, 0, 0, Width, Height, "Goto", #PB_Window_SystemMenu | #PB_Window_WindowCentered, WindowID(Main::Window\ID))
    If Window\ID
      
      Window\Text[0] = TextGadget(#PB_Any, 10, 10, 20, 20, "X:", #PB_Text_Right)
      Window\Spin[0] = SpinGadget(#PB_Any, 40, 10, 100, 20, -1000000, 1000000, #PB_Spin_Numeric)
      Window\Text[1] = TextGadget(#PB_Any, 10, 30, 20, 20, "Y:", #PB_Text_Right)
      Window\Spin[1] = SpinGadget(#PB_Any, 40, 30, 100, 20, -1000000, 1000000, #PB_Spin_Numeric)
      Window\Button = ButtonGadget(#PB_Any, 10, Height-40, Width-20, 30, "Goto")
      
      SetGadgetState(Window\Spin[0], -Main::Settings\X / Main::Settings\Zoom) : SetGadgetText(Window\Spin[0], Str(-Main::Settings\X / Main::Settings\Zoom))
      SetGadgetState(Window\Spin[1], -Main::Settings\Y / Main::Settings\Zoom) : SetGadgetText(Window\Spin[1], Str(-Main::Settings\Y / Main::Settings\Zoom))
      
      BindGadgetEvent(Window\Button, @Button_Event())
      
      BindEvent(#PB_Event_SizeWindow, @Event_SizeWindow(), Window\ID)
      BindEvent(#PB_Event_Menu, @Event_Menu(), Window\ID)
      BindEvent(#PB_Event_CloseWindow, @Event_CloseWindow(), Window\ID)
      
      ProcedureReturn #True
    EndIf
    
    ProcedureReturn #False
  EndProcedure
  
  Procedure Close()
    If Window\ID
      
      UnbindGadgetEvent(Window\Button, @Button_Event())
      
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
; CursorPosition = 127
; FirstLine = 117
; Folding = --
; EnableXP