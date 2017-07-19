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

DeclareModule Templates
  EnableExplicit
  
  ; ##################################################### Constants #################################################
  Enumeration
    #Reorder_None
    
    #Reorder_Randomize
    
    #Reorder_Inside_First_Circle
    #Reorder_Inside_First_Square
    #Reorder_Outside_First_Circle
    #Reorder_Outside_First_Square
    
    #Reorder_Rarest_Colors_First
    #Reorder_Rarest_Colors_Last
    #Reorder_Biggest_Colordifference_First
    #Reorder_Smallest_Colordifference_First
    
    #Reorder_Center_First
    #Reorder_Center_Last
    
    #Reorder_Left_To_Right
    #Reorder_Right_To_Left
    #Reorder_Top_To_Bottom
    #Reorder_Bottom_To_Top
    
    #Reorder_Replace_White_First
    #Reorder_Replace_White_Last
  EndEnumeration
  
  Enumeration
    #Draw_Result_Success
    #Draw_Result_Nothing_To_Draw
    #Draw_Result_Error
  EndEnumeration
  
  #Reorder_Amount = 4
  
  #Regression_Amount = 30
  
  ; ##################################################### Structures ################################################
  Structure Object_Settings
    Active.i
    
    Filename.s
    X.i
    Y.i
    Reorder.i [#Reorder_Amount]
    
    Correct.i
    Total.i
    Counter.i
    Total_Time.d              ; Total time in milliseconds
  EndStructure
  
  Structure Difference
    X.i
    Y.i
    
    Canvas_Color.i
    Template_Color.i
    
    Canvas_Color_Index.i
    Template_Color_Index.i
    
    Reorder_Temp.d
  EndStructure
  
  Structure Change
    Count.i
    
    Timestamp.q
  EndStructure
  
  Structure Object
    Settings.Object_Settings
    
    Template_ImageID.i    ; Template image
    Canvas_ImageID.i      ; Current snapshot from canvas
    
    File_Date.i
    
    Width.i
    Height.i
    
    List Difference.Difference()
    Recalculate_Differences.i
    
    List Change.Change()
    Rate.d
  EndStructure
  
  Structure Window
    ID.i
    Close.l
    
    ; #### Gadgets
    ListIcon.i
    Text.i [10]
    String.i
    Spin.i [10]
    Button.i [10]
    Text_Reorder.i [#Reorder_Amount]
    ComboBox_Reorder.i [#Reorder_Amount]
  EndStructure
  Global Window.Window
  
  ; ##################################################### Variables #################################################
  Global NewList Object.Object()
  
  ; ##################################################### Functions #################################################
  Declare   Settings_Save(Filename.s)
  Declare   Settings_Load(Filename.s)
  
  Declare   Window_Open()
  
  Declare   Update_Pixel(X.i, Y.i, Color.l)
  Declare   Update_Chunk(*Chunk.Main::Chunk)
  
  Declare   Main()
  
EndDeclareModule

; ###################################################################################################################
; ##################################################### Private #####################################################
; ###################################################################################################################

Module Templates
  UseModule Helper
  
  ; ##################################################### Includes ####################################################
  
  ; ##################################################### Prototypes ##################################################
  
  ; ##################################################### Structures ##################################################
  
  ; ##################################################### Constants ###################################################
  
  ; ##################################################### Structures ##################################################
  Structure Main
    
  EndStructure
  Global Main.Main
  
  ; ##################################################### Variables ###################################################
  
  ; ##################################################### Icons ... ###################################################
  
  ; ##################################################### Init ########################################################
  
  ; ##################################################### Declares ####################################################
  Declare   Refresh_ListIcon()
  Declare   Refresh_Fields()
  
  ; ##################################################### Procedures ##################################################
  Procedure Object_Delete(*Object.Object)
    ChangeCurrentElement(Object(), *Object)
    
    If Object()\Template_ImageID
      FreeImage(Object()\Template_ImageID) : Object()\Template_ImageID = 0
    EndIf
    
    If Object()\Canvas_ImageID
      FreeImage(Object()\Canvas_ImageID) : Object()\Canvas_ImageID = 0
    EndIf
    
    DeleteElement(Object())
  EndProcedure
  
  Procedure Settings_Save(Filename.s)
    Protected JSON_Array, JSON_Element
    
    Protected JSON = CreateJSON(#PB_Any)
    If JSON
      JSON_Array = SetJSONArray(JSONValue(JSON))
      
      ForEach Object()
        JSON_Element = AddJSONElement(JSON_Array)
        InsertJSONStructure(JSON_Element, Object()\Settings, Object_Settings)
      Next
      
      SaveJSON(JSON, Filename, #PB_JSON_PrettyPrint)
      FreeJSON(JSON)
    EndIf
    
  EndProcedure
  
  Procedure Settings_Load(Filename.s)
    Protected i
    Protected JSON_Element
    
    While FirstElement(Object())
      Object_Delete(Object())
    Wend
    
    Protected JSON = LoadJSON(#PB_Any, Filename)
    If JSON
      
      For i = 0 To JSONArraySize(JSONValue(JSON)) - 1
        JSON_Element = GetJSONElement(JSONValue(JSON), i)
        
        AddElement(Object())
        
        ExtractJSONStructure(JSON_Element, Object()\Settings, Object_Settings)
      Next
      
      FreeJSON(JSON)
    EndIf
    
    Refresh_ListIcon()
    Refresh_Fields()
    
  EndProcedure
  
  Procedure Refresh_ListIcon()
    Protected Progress.d
    
    If Not Window\ID
      ProcedureReturn
    EndIf
    
    While CountGadgetItems(Window\ListIcon) > ListSize(Object())
      RemoveGadgetItem(Window\ListIcon, 0)
    Wend
    While CountGadgetItems(Window\ListIcon) < ListSize(Object())
      AddGadgetItem(Window\ListIcon, -1, "")
    Wend
    
    PushListPosition(Object())
    ForEach Object()
      If Object()\Settings\Total > 0
        Progress = Object()\Settings\Correct / Object()\Settings\Total
      Else
        Progress = NaN()
      EndIf
      SetGadgetItemText(Window\ListIcon, ListIndex(Object()), Object()\Settings\Filename, 0)
      SetGadgetItemText(Window\ListIcon, ListIndex(Object()), Str(Object()\Settings\X), 1)
      SetGadgetItemText(Window\ListIcon, ListIndex(Object()), Str(Object()\Settings\Y), 2)
      SetGadgetItemText(Window\ListIcon, ListIndex(Object()), StrD(Progress * 100, 2) + "% ("+Str(Object()\Settings\Correct)+"/"+Str(Object()\Settings\Total)+")", 3)
      SetGadgetItemText(Window\ListIcon, ListIndex(Object()), Str(Object()\Settings\Counter), 4)
      SetGadgetItemText(Window\ListIcon, ListIndex(Object()), StrD(Object()\Settings\Total_Time / 3600000, 2) + "h", 5)
      SetGadgetItemText(Window\ListIcon, ListIndex(Object()), StrD((Object()\Settings\Total - Object()\Settings\Correct) / Object()\Rate / 3600, 2) + "h (" + StrD(Object()\Rate * 3600, 2) + "/h)", 6)
      If Object()\Settings\Active
        SetGadgetItemState(Window\ListIcon, ListIndex(Object()), GetGadgetItemState(Window\ListIcon, ListIndex(Object())) | #PB_ListIcon_Checked)
      Else
        SetGadgetItemState(Window\ListIcon, ListIndex(Object()), GetGadgetItemState(Window\ListIcon, ListIndex(Object())) & ~#PB_ListIcon_Checked)
      EndIf
    Next
    PopListPosition(Object())
  EndProcedure
  
  Procedure Refresh_Fields()
    Protected i
    If Not Window\ID
      ProcedureReturn
    EndIf
    
    Protected Index = GetGadgetState(Window\ListIcon)
    
    If Index <> -1 And SelectElement(Object(), Index)
      SetGadgetText(Window\String, Object()\Settings\Filename)
      SetGadgetState(Window\Spin[0], Object()\Settings\X)
      SetGadgetState(Window\Spin[1], Object()\Settings\Y)
      For i = 0 To #Reorder_Amount-1
        SetGadgetState(Window\ComboBox_Reorder[i], Object()\Settings\Reorder[i])
      Next
      
      DisableGadget(Window\String, #False)
      DisableGadget(Window\Spin[0], #False)
      DisableGadget(Window\Spin[1], #False)
      For i = 0 To #Reorder_Amount-1
        DisableGadget(Window\ComboBox_Reorder[i], #False)
      Next
      
      DisableGadget(Window\Button[0], #False)
      DisableGadget(Window\Button[1], #False)
      DisableGadget(Window\Button[2], #False)
    Else
      SetGadgetText(Window\String, "")
      SetGadgetState(Window\Spin[0], 0)
      SetGadgetState(Window\Spin[1], 0)
      For i = 0 To #Reorder_Amount-1
        SetGadgetState(Window\ComboBox_Reorder[i], 0)
      Next
      
      DisableGadget(Window\String, #True)
      DisableGadget(Window\Spin[0], #True)
      DisableGadget(Window\Spin[1], #True)
      For i = 0 To #Reorder_Amount-1
        DisableGadget(Window\ComboBox_Reorder[i], #True)
      Next
      DisableGadget(Window\Button[0], #True)
      DisableGadget(Window\Button[1], #True)
      DisableGadget(Window\Button[2], #True)
    EndIf
  EndProcedure
  
  Procedure Save_Fields()
    Protected i
    If Not Window\ID
      ProcedureReturn
    EndIf
    
    Protected Index = GetGadgetState(Window\ListIcon)
    
    If Index <> -1 And SelectElement(Object(), Index)
      Object()\Settings\Filename = GetGadgetText(Window\String)
      Object()\Settings\X = GetGadgetState(Window\Spin[0])
      Object()\Settings\Y = GetGadgetState(Window\Spin[1])
      For i = 0 To #Reorder_Amount-1
        Object()\Settings\Reorder[i] = GetGadgetState(Window\ComboBox_Reorder[i])
      Next
      ; #### Canvas image is not valid anymore, if the position was changed
      If Object()\Canvas_ImageID
        FreeImage(Object()\Canvas_ImageID) : Object()\Canvas_ImageID = 0
        ClearList(Object()\Difference())
      EndIf
    EndIf
    
    Main::Window\Canvas\Redraw = #True
  EndProcedure
  
  Procedure Event_ListIcon()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    Select Event_Type
      Case #PB_EventType_LeftClick
        Refresh_Fields()
        
      Case #PB_EventType_Change
        ForEach Object()
          If GetGadgetItemState(Event_Gadget, ListIndex(Object())) & #PB_ListIcon_Checked
            Object()\Settings\Active = #True
          Else
            Object()\Settings\Active = #False
          EndIf
        Next
        
    EndSelect
    
  EndProcedure
  
  Procedure Event_String()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    Select Event_Type
      Case #PB_EventType_Change
        Save_Fields()
        Refresh_ListIcon()
        
    EndSelect
    
  EndProcedure
  
  Procedure Event_Spin()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    Select Event_Type
      Case #PB_EventType_Change
        Save_Fields()
        Refresh_ListIcon()
        
    EndSelect
    
  EndProcedure
  
  Procedure Event_ComboBox_Reorder()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    Protected i
    For i = 0 To #Reorder_Amount-1
      If Event_Gadget = Window\ComboBox_Reorder[i]
        
        Select Event_Type
          Case #PB_EventType_Change
            Save_Fields()
            Refresh_ListIcon()
            
        EndSelect
        
        Break
      EndIf
    Next
    
  EndProcedure
  
  Procedure Event_Button()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    Protected Index = GetGadgetState(Window\ListIcon)
    
    Select Event_Gadget
      Case Window\Button[0]
        If Index > 0
          SwapElements(Object(), SelectElement(Object(), Index), SelectElement(Object(), Index-1))
          
          SetGadgetState(Window\ListIcon, GetGadgetState(Window\ListIcon) - 1)
          
          Refresh_ListIcon()
          Refresh_Fields()
        EndIf
        
      Case Window\Button[1]
        If Index < ListSize(Object()) - 1
          SwapElements(Object(), SelectElement(Object(), Index), SelectElement(Object(), Index+1))
          
          SetGadgetState(Window\ListIcon, GetGadgetState(Window\ListIcon) + 1)
          
          Refresh_ListIcon()
          Refresh_Fields()
        EndIf
        
      Case Window\Button[2]
        If Index <> -1 And SelectElement(Object(), Index)
          
          DeleteElement(Object())
          
          Refresh_ListIcon()
          Refresh_Fields()
        EndIf
        
      Case Window\Button[3]
        LastElement(Object())
        AddElement(Object())
        
        Object()\Settings\Reorder[0] = #Reorder_Randomize
        Object()\Settings\Reorder[1] = #Reorder_Inside_First_Square
        Object()\Settings\Reorder[2] = #Reorder_Biggest_Colordifference_First
        
        Refresh_ListIcon()
        Refresh_Fields()
        
    EndSelect
  EndProcedure
  
  Procedure Event_SizeWindow()
    Protected Event_Window = EventWindow()
    Protected Event_Gadget = EventGadget()
    Protected Event_Type = EventType()
    
    Protected X, Y, i
    
    Protected Width = WindowWidth(Event_Window)
    Protected Height = WindowHeight(Event_Window)
    
    X = 0
    
    ResizeGadget(Window\ListIcon, 0, 0, Width, Height - 110 - #Reorder_Amount*20) : Y + (Height - 110 - #Reorder_Amount*20)
    
    Y + 10
    
    ResizeGadget(Window\Text [0], 10, Y, 100, 20)
    ResizeGadget(Window\Spin [0], 120, Y, 100, 20) : Y + 20
    ResizeGadget(Window\Text [1], 10, Y, 100, 20)
    ResizeGadget(Window\Spin [1], 120, Y, 100, 20) : Y + 20
    ResizeGadget(Window\Text [2], 10, Y, 100, 20)
    ResizeGadget(Window\String, 120, Y, Width - 130, 20) : Y + 20
    
    For i = 0 To #Reorder_Amount-1
      ResizeGadget(Window\Text_Reorder [i], 10, Y, 100, 20)
      ResizeGadget(Window\ComboBox_Reorder [i], 120, Y, Width - 130, 20) : Y + 20
    Next
    
    Y + 10
    
    ResizeGadget(Window\Button [0], 10, Y, 50, 20)
    ResizeGadget(Window\Button [1], 60, Y, 50, 20)
    ResizeGadget(Window\Button [2], 110, Y, 50, 20)
    ResizeGadget(Window\Button [3], 160, Y, 50, 20)
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
    Window\Close = #True
  EndProcedure
  
  Procedure Window_Open()
    Protected Width, Height
    Protected X, Y
    Protected i
    
    If Window\ID = 0
      
      Width = 650
      Height = 500
      X = 0
      
      Window\ID = OpenWindow(#PB_Any, 0, 0, Width, Height, "Templates", #PB_Window_SystemMenu | #PB_Window_WindowCentered | #PB_Window_SizeGadget, WindowID(Main::Window\ID))
      
      Window\ListIcon = ListIconGadget(#PB_Any, 0, 0, Width, Height - 110 - #Reorder_Amount*20, "File", 150, #PB_ListIcon_CheckBoxes | #PB_ListIcon_GridLines | #PB_ListIcon_FullRowSelect | #PB_ListIcon_AlwaysShowSelection) : Y + (Height - 110 - #Reorder_Amount*20)
      AddGadgetColumn(Window\ListIcon, 1, "X", 50)
      AddGadgetColumn(Window\ListIcon, 2, "Y", 50)
      AddGadgetColumn(Window\ListIcon, 3, "Progress", 120)
      AddGadgetColumn(Window\ListIcon, 4, "Counter", 60)
      AddGadgetColumn(Window\ListIcon, 5, "Time", 80)
      AddGadgetColumn(Window\ListIcon, 6, "ETA", 120)
      
      Y + 10
      
      Window\Text [0] = TextGadget(#PB_Any, 10, Y, 100, 20, "X:", #PB_Text_Right)
      Window\Spin [0] = SpinGadget(#PB_Any, 120, Y, 100, 20, -10000, 10000, #PB_Spin_Numeric) : Y + 20
      Window\Text [1] = TextGadget(#PB_Any, 10, Y, 100, 20, "Y:", #PB_Text_Right)
      Window\Spin [1] = SpinGadget(#PB_Any, 120, Y, 100, 20, -10000, 10000, #PB_Spin_Numeric) : Y + 20
      Window\Text [2] = TextGadget(#PB_Any, 10, Y, 100, 20, "File:", #PB_Text_Right)
      Window\String = StringGadget(#PB_Any, 120, Y, Width - 130, 20, "") : Y + 20
      For i = 0 To #Reorder_Amount-1
        Window\Text_Reorder [i] = TextGadget(#PB_Any, 10, Y, 100, 20, "Reorder Algorithm "+i+":", #PB_Text_Right)
        Window\ComboBox_Reorder [i] = ComboBoxGadget(#PB_Any, 120, Y, Width - 130, 20) : Y + 20
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_None, "None")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Randomize, "Randomize")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Inside_First_Circle, "Inside template first (Circle)")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Inside_First_Square, "Inside template first (Square)")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Outside_First_Circle, "Outside template first (Circle)")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Outside_First_Square, "Outside template first (Square)")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Rarest_Colors_First, "Rarest colors first")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Rarest_Colors_Last, "Rarest colors last")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Biggest_Colordifference_First, "Biggest colordifference first")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Smallest_Colordifference_First, "Smallest colordifference first")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Center_First, "Center first (Circle)")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Center_Last, "Center last (Circle)")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Left_To_Right, "Left to right")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Right_To_Left, "Right to left")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Top_To_Bottom, "Top to bottom")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Bottom_To_Top, "Bottom to top")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Replace_White_First, "Replace white first")
        AddGadgetItem(Window\ComboBox_Reorder [i], #Reorder_Replace_White_Last, "Replace white last")
      Next
      
      Y + 10
      
      Window\Button [0] = ButtonGadget(#PB_Any, 10, Y, 50, 20, "Up")
      Window\Button [1] = ButtonGadget(#PB_Any, 60, Y, 50, 20, "Down")
      Window\Button [2] = ButtonGadget(#PB_Any, 110, Y, 50, 20, "Delete")
      Window\Button [3] = ButtonGadget(#PB_Any, 160, Y, 50, 20, "Create")
      
      BindGadgetEvent(Window\ListIcon, @Event_ListIcon())
      BindGadgetEvent(Window\Spin [0], @Event_Spin())
      BindGadgetEvent(Window\Spin [1], @Event_Spin())
      BindGadgetEvent(Window\String, @Event_String())
      For i = 0 To #Reorder_Amount-1
        BindGadgetEvent(Window\ComboBox_Reorder[i], @Event_ComboBox_Reorder())
      Next
      BindGadgetEvent(Window\Button [0], @Event_Button())
      BindGadgetEvent(Window\Button [1], @Event_Button())
      BindGadgetEvent(Window\Button [2], @Event_Button())
      BindGadgetEvent(Window\Button [3], @Event_Button())
      
      BindEvent(#PB_Event_SizeWindow, @Event_SizeWindow(), Window\ID)
      ;BindEvent(#PB_Event_Repaint, @Event_SizeWindow(), Window\ID)
      ;BindEvent(#PB_Event_RestoreWindow, @Event_SizeWindow(), Window\ID)
      BindEvent(#PB_Event_Menu, @Event_Menu(), Window\ID)
      BindEvent(#PB_Event_CloseWindow, @Event_CloseWindow(), Window\ID)
      
      Refresh_ListIcon()
      Refresh_Fields()
      
    EndIf
  EndProcedure
  
  Procedure Window_Close()
    Protected i
    
    If Window\ID
      
      UnbindGadgetEvent(Window\ListIcon, @Event_ListIcon())
      UnbindGadgetEvent(Window\Spin [0], @Event_Spin())
      UnbindGadgetEvent(Window\Spin [1], @Event_Spin())
      UnbindGadgetEvent(Window\String, @Event_String())
      For i = 0 To #Reorder_Amount-1
        UnbindGadgetEvent(Window\ComboBox_Reorder[i], @Event_ComboBox_Reorder())
      Next
      UnbindGadgetEvent(Window\Button [0], @Event_Button())
      UnbindGadgetEvent(Window\Button [1], @Event_Button())
      UnbindGadgetEvent(Window\Button [2], @Event_Button())
      UnbindGadgetEvent(Window\Button [3], @Event_Button())
      
      UnbindEvent(#PB_Event_SizeWindow, @Event_SizeWindow(), Window\ID)
      ;UnbindEvent(#PB_Event_Repaint, @Event_SizeWindow(), Window\ID)
      ;UnbindEvent(#PB_Event_RestoreWindow, @Event_SizeWindow(), Window\ID)
      UnbindEvent(#PB_Event_Menu, @Event_Menu(), Window\ID)
      UnbindEvent(#PB_Event_CloseWindow, @Event_CloseWindow(), Window\ID)
      
      CloseWindow(Window\ID)
      Window\ID = 0
    EndIf
  EndProcedure
  
  Procedure Image_Get(Filename.s)
    Protected ImageID = LoadImage(#PB_Any, Filename)
    Protected ix, iy
    Protected Width, Height
    Protected Color_Index
    Protected No_Alpha
    
    If ImageID
      If ImageDepth(ImageID, #PB_Image_InternalDepth) = 24
        No_Alpha = #True
      Else
        No_Alpha = #False
      EndIf
      
      If StartDrawing(ImageOutput(ImageID))
        DrawingMode(#PB_2DDrawing_AllChannels)
        Width = OutputWidth()
        Height = OutputHeight()
        
        For ix = 0 To Width - 1
          For iy = 0 To Height - 1
            If Alpha(Point(ix, iy)) > 127 Or No_Alpha
              Color_Index = Main::Get_Color_Index(Point(ix, iy))
              Plot(ix, iy, Main::Palette(Color_Index)\Color)
            Else
              Plot(ix, iy, 0)
            EndIf
          Next
        Next
        
        StopDrawing()
      EndIf
    Else
      Debug "Image '"+Filename+"' couldn't be loaded"
    EndIf
    
    ProcedureReturn ImageID
  EndProcedure
  
  Procedure Reorder_Randomize(*Object.Object)
    RandomizeList(*Object\Difference())
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Reorder_Center_Distance(*Object.Object, Options=#PB_Sort_Ascending)
    ForEach *Object\Difference()
      *Object\Difference()\Reorder_Temp = Pow(*Object\Difference()\X + *Object\Settings\X - Main::Userdata\Center_X, 2) + Pow(*Object\Difference()\Y + *Object\Settings\Y - Main::Userdata\Center_Y, 2)
    Next
    
    SortStructuredList(*Object\Difference(), Options, OffsetOf(Difference\Reorder_Temp), TypeOf(Difference\Reorder_Temp))
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Reorder_Template_Center_Distance(*Object.Object, Options=#PB_Sort_Ascending)
    ForEach *Object\Difference()
      *Object\Difference()\Reorder_Temp = Pow((*Object\Width-1) / 2 - *Object\Difference()\X, 2) + Pow((*Object\Height-1) / 2 - *Object\Difference()\Y, 2)
    Next
    
    SortStructuredList(*Object\Difference(), Options, OffsetOf(Difference\Reorder_Temp), TypeOf(Difference\Reorder_Temp))
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Reorder_Template_Center_Distance_Square(*Object.Object, Options=#PB_Sort_Ascending)
    Protected Distance.d
    
    ForEach *Object\Difference()
      Distance = Abs(*Object\Difference()\X - (*Object\Width-1) / 2)
      If Distance < Abs(*Object\Difference()\Y - (*Object\Height-1) / 2)
        Distance = Abs(*Object\Difference()\Y - (*Object\Height-1) / 2)
      EndIf
      *Object\Difference()\Reorder_Temp = Distance
    Next
    
    SortStructuredList(*Object\Difference(), Options, OffsetOf(Difference\Reorder_Temp), TypeOf(Difference\Reorder_Temp))
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Reorder_Color_Amount(*Object.Object, Options=#PB_Sort_Ascending)
    Protected Dim Color(Main::#Colors-1)
    Protected First
    Protected i, Color_Index
    Protected Min
    
    ForEach *Object\Difference()
      Color(*Object\Difference()\Template_Color_Index) + 1
    Next
    
    ForEach *Object\Difference()
      *Object\Difference()\Reorder_Temp = Color(*Object\Difference()\Template_Color_Index)
    Next
    
    SortStructuredList(*Object\Difference(), Options, OffsetOf(Difference\Reorder_Temp), TypeOf(Difference\Reorder_Temp))
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Reorder_Colordifference(*Object.Object, Options=#PB_Sort_Ascending)
    ForEach *Object\Difference()
      *Object\Difference()\Reorder_Temp = Color_Distance_Squared(*Object\Difference()\Canvas_Color, *Object\Difference()\Template_Color)
    Next
    
    SortStructuredList(*Object\Difference(), Options, OffsetOf(Difference\Reorder_Temp), TypeOf(Difference\Reorder_Temp))
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Reorder_X(*Object.Object, Options=#PB_Sort_Ascending)
    ForEach *Object\Difference()
      *Object\Difference()\Reorder_Temp = *Object\Difference()\X
    Next
    
    SortStructuredList(*Object\Difference(), Options, OffsetOf(Difference\Reorder_Temp), TypeOf(Difference\Reorder_Temp))
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Reorder_Y(*Object.Object, Options=#PB_Sort_Ascending)
    ForEach *Object\Difference()
      *Object\Difference()\Reorder_Temp = *Object\Difference()\Y
    Next
    
    SortStructuredList(*Object\Difference(), Options, OffsetOf(Difference\Reorder_Temp), TypeOf(Difference\Reorder_Temp))
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Reorder_White(*Object.Object, Options=#PB_Sort_Ascending)
    ForEach *Object\Difference()
      If *Object\Difference()\Canvas_Color_Index = 0
        *Object\Difference()\Reorder_Temp = 0
      Else
        *Object\Difference()\Reorder_Temp = 1
      EndIf
    Next
    
    SortStructuredList(*Object\Difference(), Options, OffsetOf(Difference\Reorder_Temp), TypeOf(Difference\Reorder_Temp))
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Draw(*Object.Object)
    Protected i
    Protected Counter
    Protected Result
    
    ; #### Reorder difference list
    For i = 0 To #Reorder_Amount-1
      Select *Object\Settings\Reorder [i]
        Case #Reorder_None
        Case #Reorder_Randomize                       : Reorder_Randomize(*Object)
        Case #Reorder_Inside_First_Circle             : Reorder_Template_Center_Distance(*Object, #PB_Sort_Ascending)
        Case #Reorder_Inside_First_Square             : Reorder_Template_Center_Distance_Square(*Object, #PB_Sort_Ascending)
        Case #Reorder_Outside_First_Circle            : Reorder_Template_Center_Distance(*Object, #PB_Sort_Descending)
        Case #Reorder_Outside_First_Square            : Reorder_Template_Center_Distance_Square(*Object, #PB_Sort_Descending)
        Case #Reorder_Rarest_Colors_First             : Reorder_Color_Amount(*Object, #PB_Sort_Ascending)
        Case #Reorder_Rarest_Colors_Last              : Reorder_Color_Amount(*Object, #PB_Sort_Descending)
        Case #Reorder_Biggest_Colordifference_First   : Reorder_Colordifference(*Object, #PB_Sort_Descending)
        Case #Reorder_Smallest_Colordifference_First  : Reorder_Colordifference(*Object, #PB_Sort_Ascending)
        Case #Reorder_Center_First                    : Reorder_Center_Distance(*Object, #PB_Sort_Ascending)
        Case #Reorder_Center_Last                     : Reorder_Center_Distance(*Object, #PB_Sort_Descending)
        Case #Reorder_Left_To_Right                   : Reorder_X(*Object, #PB_Sort_Ascending)
        Case #Reorder_Right_To_Left                   : Reorder_X(*Object, #PB_Sort_Descending)
        Case #Reorder_Top_To_Bottom                   : Reorder_Y(*Object, #PB_Sort_Ascending)
        Case #Reorder_Bottom_To_Top                   : Reorder_Y(*Object, #PB_Sort_Descending)
        Case #Reorder_Replace_White_First             : Reorder_White(*Object, #PB_Sort_Ascending)
        Case #Reorder_Replace_White_Last              : Reorder_White(*Object, #PB_Sort_Descending)
      EndSelect
    Next
    
    ForEach *Object\Difference()
      Result = Main::HTTP_Post_Input(*Object\Settings\X + *Object\Difference()\X, *Object\Settings\Y + *Object\Difference()\Y, *Object\Difference()\Template_Color_Index, *Object, Main::Settings\Fingerprint)
      Select Result
        Case Main::#Input_Result_Success
          ProcedureReturn #Draw_Result_Success
        Case Main::#Input_Result_Local_Error
          Counter + 1
          If Counter > 20
            ProcedureReturn #Draw_Result_Error
          EndIf
        Case Main::#Input_Result_Global_Error
          ProcedureReturn #Draw_Result_Error
      EndSelect
    Next
    
    ProcedureReturn #Draw_Result_Nothing_To_Draw
  EndProcedure
  
  Procedure Compare_Template(*Object.Object)
    Protected ix, iy
    Protected Width, Height
    Protected Total, Correct
    Protected No_Alpha
    
    If *Object\Template_ImageID And StartDrawing(ImageOutput(*Object\Template_ImageID))
      DrawingMode(#PB_2DDrawing_AllChannels)
      Width = OutputWidth()
      Height = OutputHeight()
      
      If ImageDepth(*Object\Template_ImageID, #PB_Image_InternalDepth) = 24
        No_Alpha = #True
      Else
        No_Alpha = #False
      EndIf
      
      Protected Dim Template_Array.l(Width-1, Height-1)
      
      For ix = 0 To Width - 1
        For iy = 0 To Height - 1
          If No_Alpha
            Template_Array(ix, iy) = Point(ix, iy) | $FF000000
          Else
            Template_Array(ix, iy) = Point(ix, iy)
          EndIf
        Next
      Next
      
      StopDrawing()
    Else
      Debug "*Object\Template_ImageID or StartDrawing(ImageOutput(*Object\Template_ImageID)) failed"
      ProcedureReturn #False
    EndIf
    
    If Object()\Canvas_ImageID And StartDrawing(ImageOutput(Object()\Canvas_ImageID))
      DrawingMode(#PB_2DDrawing_AllChannels)
      Width = OutputWidth()
      Height = OutputHeight()
      
      ClearList(*Object\Difference())
      
      For ix = 0 To Width - 1
        For iy = 0 To Height - 1
          If Alpha(Template_Array(ix, iy)) > 0
            Total + 1
            If Alpha(Point(ix, iy)) > 0
              If Template_Array(ix, iy) & $FFFFFF <> Point(ix, iy) & $FFFFFF
                AddElement(*Object\Difference())
                *Object\Difference()\X = ix
                *Object\Difference()\Y = iy
                *Object\Difference()\Canvas_Color = Point(ix, iy)
                *Object\Difference()\Template_Color = Template_Array(ix, iy)
                *Object\Difference()\Canvas_Color_Index = Main::Get_Color_Index(*Object\Difference()\Canvas_Color)
                *Object\Difference()\Template_Color_Index = Main::Get_Color_Index(*Object\Difference()\Template_Color)
                ;Debug "Difference found: X:" + Str(ix) + " Y:" + Str(iy)
              Else
                Correct + 1
              EndIf
            EndIf
          EndIf
        Next
      Next
      
      StopDrawing()
    Else
      Debug "Object()\Canvas_ImageID And StartDrawing(ImageOutput(Object()\Canvas_ImageID)) failed"
      ProcedureReturn #False
    EndIf
    
    ; #### Calculate the rate
    If Correct <> *Object\Settings\Correct And Correct > 0
      LastElement(*Object\Change())
      AddElement(*Object\Change())
      *Object\Change()\Count = Correct
      *Object\Change()\Timestamp = ElapsedMilliseconds()
    EndIf
    
    While ListSize(*Object\Change()) > #Regression_Amount
      FirstElement(*Object\Change())
      DeleteElement(*Object\Change())
    Wend
    
    If FirstElement(*Object\Change())
      Protected Timestamp_Start.q = *Object\Change()\Timestamp
      *Object\Rate = -*Object\Change()\Count
    EndIf
    If LastElement(*Object\Change())
      Protected Timestamp_End.q = *Object\Change()\Timestamp
      *Object\Rate + *Object\Change()\Count
    EndIf
    
    If Timestamp_End > Timestamp_Start
      *Object\Rate / (Timestamp_End - Timestamp_Start) * 1000
    EndIf
    
    *Object\Settings\Correct = Correct
    *Object\Settings\Total = Total
    
    Refresh_ListIcon()
    
    ProcedureReturn #True
  EndProcedure
  
  Procedure Update_Pixel(X.i, Y.i, Color.l)
    Protected OX, OY
    
    ForEach Object()
      OX = Object()\Settings\X
      OY = Object()\Settings\Y
      If X < OX + Object()\Width And Y < OY + Object()\Height And X >= OX And Y >= OY
        
        If Object()\Canvas_ImageID And StartDrawing(ImageOutput(Object()\Canvas_ImageID))
          DrawingMode(#PB_2DDrawing_AllChannels)
          
          Plot(X - OX, Y - OY, Color)
          
          StopDrawing()
          
          Object()\Recalculate_Differences = #True
        EndIf
      EndIf
    Next
  EndProcedure
  
  Procedure Update_Chunk(*Chunk.Main::Chunk)
    Protected R_X, R_Y
    Protected Width, Height
    Protected X, Y
    
    ForEach Object()
      X = Object()\Settings\X
      Y = Object()\Settings\Y
      Width = Object()\Width
      Height = Object()\Height
      R_X = *Chunk\CX * Main::#Chunk_Size - X
      R_Y = *Chunk\CY * Main::#Chunk_Size - Y
      If *Chunk\Image And R_X + Main::#Chunk_Size > 0 And R_X < Width And R_Y + Main::#Chunk_Size > 0 And R_Y < Height
        
        If Object()\Canvas_ImageID And StartDrawing(ImageOutput(Object()\Canvas_ImageID))
          DrawingMode(#PB_2DDrawing_AllChannels)
          
          DrawImage(ImageID(*Chunk\Image), R_X, R_Y)
          
          StopDrawing()
          
          Object()\Recalculate_Differences = #True
        EndIf
      EndIf
    Next
  EndProcedure
  
  Procedure Main()
    Static Timer_Load_Image
    Static Timer_Draw
    Static Timer_Compare
    Static Timer_Settings
    
    ; #### Load images
    If Timer_Load_Image < ElapsedMilliseconds()
      Timer_Load_Image = ElapsedMilliseconds() + 1000
      ForEach Object()
        If Object()\Settings\Filename
          If Not Object()\Template_ImageID Or Object()\File_Date <> GetFileDate(Object()\Settings\Filename, #PB_Date_Modified)
            If Object()\Template_ImageID
              FreeImage(Object()\Template_ImageID) : Object()\Template_ImageID = 0
            EndIf
            Object()\File_Date = GetFileDate(Object()\Settings\Filename, #PB_Date_Modified)
            Object()\Template_ImageID = Image_Get(Object()\Settings\Filename)
            If Object()\Template_ImageID
              If Object()\Width <> ImageWidth(Object()\Template_ImageID) Or Object()\Height <> ImageHeight(Object()\Template_ImageID)
                Object()\Width = ImageWidth(Object()\Template_ImageID)
                Object()\Height = ImageHeight(Object()\Template_ImageID)
                If Object()\Canvas_ImageID
                  FreeImage(Object()\Canvas_ImageID) : Object()\Canvas_ImageID = 0
                EndIf
              EndIf
            EndIf
            Object()\Recalculate_Differences = #True
          EndIf
        EndIf
      Next
    EndIf
    
    ; #### Load canvas image
    ForEach Object()
      If Not Object()\Canvas_ImageID And Object()\Width And Object()\Height
        Object()\Canvas_ImageID = Main::Image_Get(Object()\Settings\X, Object()\Settings\Y, Object()\Width, Object()\Height)
        Object()\Recalculate_Differences = #True
      EndIf
    Next
    
    ; #### Calculate the differences
    ForEach Object()
      If Object()\Recalculate_Differences
        Object()\Recalculate_Differences = #False
        Compare_Template(Object())
      EndIf
    Next
    
    ; #### Draw templates
    If Main::Userdata\Timestamp_Next_Pixel + 0 < Main::Get_Timestamp() And Timer_Draw < ElapsedMilliseconds()
      Timer_Draw = ElapsedMilliseconds() + 4000
      ForEach Object()
        If Object()\Settings\Active
          Select Draw(Object())
            Case #Draw_Result_Success, #Draw_Result_Error
              Break
            Case #Draw_Result_Nothing_To_Draw
              
          EndSelect
        EndIf
      Next
    EndIf
    
    ; #### Save templates every 10 seconds
    If Timer_Settings < ElapsedMilliseconds()
      Timer_Settings = ElapsedMilliseconds() + 10000
      Settings_Save(Main::Main\Path_AppData + Main::#Filename_Templates)
    EndIf
    
    If Window\ID
      If Window\Close
        Window\Close = #False
        Window_Close()
      EndIf
    EndIf
    
  EndProcedure
  
  ; ##################################################### Initialisation ##############################################
  
  ; ##################################################### Data Sections ###############################################
  
EndModule

; IDE Options = PureBasic 5.60 (Windows - x64)
; CursorPosition = 758
; FirstLine = 756
; Folding = ------
; EnableXP
; Executable = ..\Pixelcanvas Client.exe
; DisableDebugger
; EnableCompileCount = 0
; EnableBuildCount = 0
; EnableExeConstant