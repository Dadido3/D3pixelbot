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

DeclareModule About
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

Module About
  UsePNGImageDecoder()
  
  ; ##################################################### Includes ####################################################
  
  ; ##################################################### Prototypes ##################################################
  
  ; ##################################################### Structures ##################################################
  
  ; ##################################################### Constants ###################################################
  
  ; ##################################################### Structures ##################################################
  
  Structure Main
    
  EndStructure
  Global Main.Main
  
  Structure About
    Window_ID.i
    Window_Close.l
    
    ; #### Gadgets
    Canvas.i
    Editor.i
    
    Redraw.l
  EndStructure
  Global About.About
  
  ; ##################################################### Variables ###################################################
  
  ; ##################################################### Icons ... ###################################################
  
  ; ##################################################### Init ########################################################
  
  Global Image_Logo = CatchImage(#PB_Any, ?Image_Logo)
  
  Global Font = LoadFont(#PB_Any, "Courier New", 10)
  
  ; ##################################################### Procedures ##################################################
  
  Procedure Canvas_Redraw()
    Protected Width = GadgetWidth(About\Canvas)
    Protected Height = GadgetHeight(About\Canvas)
    Protected Text.s = "V. "+StrF(Main::#Version*0.001, 3)
    
    If StartDrawing(CanvasOutput(About\Canvas))
      
      Box(0, 0, Width, Height, RGB(220,220,220))
      
      DrawImage(ImageID(Image_Logo), 0, 0, Width, Height)
      
      DrawingMode(#PB_2DDrawing_Transparent)
      DrawingFont(FontID(Font))
      DrawText(Width-TextWidth(Text), Height-TextHeight(Text), Text, RGB(255,255,255))
      
      StopDrawing()
    EndIf
  EndProcedure
  
  Procedure Editor_Fill()
    Protected Description.s, Temp_Text.s
    
    ClearGadgetItems(About\Editor)
    
    SetGadgetFont(About\Editor, FontID(Font))
    
    AddGadgetItem(About\Editor, -1, "                 ╔══════════════════════════╗")
    AddGadgetItem(About\Editor, -1, "                 ║  Pixelcanvas Cl. V"+StrF(Main::#Version*0.001, 3)+"  ║")
    AddGadgetItem(About\Editor, -1, "                 ╟──────────────────────────╢")
    AddGadgetItem(About\Editor, -1, "                 ║  Custom Pixelcanvas Cl.  ║")
    AddGadgetItem(About\Editor, -1, "                 ╚══════════════════════════╝")
    AddGadgetItem(About\Editor, -1, "")
    AddGadgetItem(About\Editor, -1, "Programmer: David Vogel (Dadido3, Xaardas)")
    AddGadgetItem(About\Editor, -1, "Website: www.D3nexus.de")
    AddGadgetItem(About\Editor, -1, "Repository: Not public yet");www.github.com/Dadido3/...")
    AddGadgetItem(About\Editor, -1, "")
    AddGadgetItem(About\Editor, -1, "╔══════════════════════╗")
    AddGadgetItem(About\Editor, -1, "║     Compilation:     ║")
    AddGadgetItem(About\Editor, -1, "╚══════════════════════╝")
    AddGadgetItem(About\Editor, -1, "")
    AddGadgetItem(About\Editor, -1, "Times compiled:  "+Str(#PB_Editor_CompileCount))
    AddGadgetItem(About\Editor, -1, "Times built:     "+Str(#PB_Editor_BuildCount))
    AddGadgetItem(About\Editor, -1, "Build timestamp: "+FormatDate("%hh:%ii:%ss %dd.%mm.%yyyy", #PB_Compiler_Date))
    CompilerIf #PB_Compiler_Processor = #PB_Processor_x86
      AddGadgetItem(About\Editor, -1, "Compiled with PureBasic "+StrF(#PB_Compiler_Version/100, 2)+" (x86)")
    CompilerElse
      AddGadgetItem(About\Editor, -1, "Compiled with PureBasic "+StrF(#PB_Compiler_Version/100, 2)+" (x64)")
    CompilerEndIf
    AddGadgetItem(About\Editor, -1, "")
    AddGadgetItem(About\Editor, -1, "╔══════════════════════╗")
    AddGadgetItem(About\Editor, -1, "║      Thanks to:      ║")
    AddGadgetItem(About\Editor, -1, "╚══════════════════════╝")
    AddGadgetItem(About\Editor, -1, "")
    AddGadgetItem(About\Editor, -1, "► Silk icons from http://www.famfamfam.com/lab/icons/silk/")
    AddGadgetItem(About\Editor, -1, "")
  EndProcedure
  
  Procedure Event_Canvas()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    
  EndProcedure
  
  Procedure Event_SizeWindow()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    
    About\Redraw = #True
  EndProcedure
  
  Procedure Event_ActivateWindow()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    About\Redraw = #True
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
    
    ;Close()
    About\Window_Close = #True
  EndProcedure
  
  Procedure Open()
    Protected Width, Height
    
    If About\Window_ID = 0
      
      Width = 500
      Height = 600
      
      About\Window_ID = OpenWindow(#PB_Any, 0, 0, Width, Height, "About", #PB_Window_SystemMenu | #PB_Window_WindowCentered, WindowID(Main::Window\ID))
      
      About\Canvas = CanvasGadget(#PB_Any, 0, 0, Width, 279)
      About\Editor = EditorGadget(#PB_Any, 0, 279, Width, Height-279, #PB_Editor_ReadOnly | #PB_Editor_WordWrap)
      
      Editor_Fill()
      
      BindGadgetEvent(About\Canvas, @Event_Canvas())
      
      BindEvent(#PB_Event_SizeWindow, @Event_SizeWindow(), About\Window_ID)
      ;BindEvent(#PB_Event_Repaint, @Event_SizeWindow(), About\Window_ID)
      ;BindEvent(#PB_Event_RestoreWindow, @Event_SizeWindow(), About\Window_ID)
      BindEvent(#PB_Event_Menu, @Event_Menu(), About\Window_ID)
      BindEvent(#PB_Event_CloseWindow, @Event_CloseWindow(), About\Window_ID)
      
      About\Redraw = #True
      
    EndIf
  EndProcedure
  
  Procedure Close()
    If About\Window_ID
      
      UnbindGadgetEvent(About\Canvas, @Event_Canvas())
      
      UnbindEvent(#PB_Event_SizeWindow, @Event_SizeWindow(), About\Window_ID)
      ;UnbindEvent(#PB_Event_Repaint, @Event_SizeWindow(), About\Window_ID)
      ;UnbindEvent(#PB_Event_RestoreWindow, @Event_SizeWindow(), About\Window_ID)
      UnbindEvent(#PB_Event_Menu, @Event_Menu(), About\Window_ID)
      UnbindEvent(#PB_Event_CloseWindow, @Event_CloseWindow(), About\Window_ID)
      
      CloseWindow(About\Window_ID)
      About\Window_ID = 0
    EndIf
  EndProcedure
  
  Procedure Main()
    If Not About\Window_ID
      ProcedureReturn #False
    EndIf
    
    If About\Redraw
      About\Redraw = #False
      Canvas_Redraw()
    EndIf
    
    If About\Window_Close
      About\Window_Close = #False
      Close()
    EndIf
    
  EndProcedure
  
  ; ##################################################### Initialisation ##############################################
  
  
  
  ; ##################################################### Data Sections ###############################################
  
  DataSection
    Image_Logo:
    IncludeBinary "../Data/Images/Logo.png"
  EndDataSection
  
EndModule

; IDE Options = PureBasic 5.60 beta 6 (Windows - x64)
; CursorPosition = 32
; FirstLine = 8
; Folding = --
; EnableXP