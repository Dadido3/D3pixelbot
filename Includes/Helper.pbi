; ##################################################### License / Copyright #########################################
; 
; ##################################################### Documentation / Comments ####################################
; 
; 
; 
; 
; 
; 
; 
; 
; ###################################################################################################################
; ##################################################### Public ######################################################
; ###################################################################################################################

DeclareModule Helper
  EnableExplicit
  ; ################################################### Constants ###################################################
  
  ; ################################################### Macros ######################################################
  Macro Line(x, y, Width, Height, Color)
    LineXY((x), (y), (x)+(Width), (y)+(Height), (Color))
  EndMacro
  
  ; ################################################### Functions ###################################################
  Declare.s GetFullPathName(Filename.s)
  Declare   IsChildOfPath(Parent.s, Child.s)
  Declare.s SHGetFolderPath(CSIDL)
  Declare   MakeSureDirectoryPathExists(Path.s)
  
  Declare.q Quad_Divide_Floor(A.q, B.q)
  Declare.q Quad_Divide_Ceil(A.q, B.q)
  
  Declare.d Color_Distance(Color_A.l, Color_B.l)
  Declare.d Color_Distance_Squared(Color_A.l, Color_B.l)
  
EndDeclareModule

; ###################################################################################################################
; ##################################################### Private #####################################################
; ###################################################################################################################

Module Helper
  EnableExplicit
  ; ################################################### Structures ##################################################
  
  ; ################################################### Libraries ###################################################
  
  ; ################################################### Procedures ##################################################
  Procedure.s GetFullPathName(Filename.s)
    Protected Characters
    Protected *Temp_Buffer
    Protected Result.s
    
    Characters = GetFullPathName_(@Filename, #Null, #Null, #Null)
    *Temp_Buffer = AllocateMemory(Characters * SizeOf(Character))
    
    GetFullPathName_(@Filename, Characters, *Temp_Buffer, #Null)
    Result = PeekS(*Temp_Buffer, Characters)
    
    FreeMemory(*Temp_Buffer)
    
    ProcedureReturn Result
  EndProcedure
  
  Procedure IsChildOfPath(Parent.s, Child.s)
    Protected Parent_Full.s = GetPathPart(GetFullPathName(Parent))
    Protected Child_Full.s = GetPathPart(GetFullPathName(Child))
    
    If Left(Child_Full, Len(Parent_Full)) = Parent_Full
      ProcedureReturn #True
    Else
      ProcedureReturn #False
    EndIf
  EndProcedure
  
  Procedure.s SHGetFolderPath(CSIDL)
    Protected *String = AllocateMemory(#MAX_PATH+1)
    SHGetFolderPath_(0, CSIDL, #Null, 0, *String)     ; Doesn't include the last "\"
    Protected String.s = PeekS(*String)
    FreeMemory(*String)
    ProcedureReturn String
  EndProcedure
  
  Procedure MakeSureDirectoryPathExists(Path.s)
    Protected Parent_Path.s
    Path = GetPathPart(Path)
    Path = ReplaceString(Path, "\", "/")
    
    If FileSize(Path) = -2
      ; #### Directory exists
      ProcedureReturn #True
    Else
      ; #### Directory doesn't exist. Check (and create) parent directory, and then create the final directory
      Parent_Path = ReverseString(Path)
      Parent_Path = RemoveString(Parent_Path, "/", #PB_String_CaseSensitive, 1, 1)
      Parent_Path = Mid(Parent_Path, FindString(Parent_Path, "/"))
      Parent_Path = ReverseString(Parent_Path)
      If MakeSureDirectoryPathExists(Parent_Path)
        CreateDirectory(Path)
        ProcedureReturn #True
      EndIf
    EndIf
    
    ProcedureReturn #False
  EndProcedure
  
  ; #### Works perfectly, A and B can be positive or negative. B must not be zero!
  Procedure.q Quad_Divide_Floor(A.q, B.q)
    Protected Temp.q = A / B
    If (((a ! b) < 0) And (a % b <> 0))
      ProcedureReturn Temp - 1
    Else
      ProcedureReturn Temp
    EndIf
  EndProcedure
  
  ; #### Works perfectly, A and B can be positive or negative. B must not be zero!
  Procedure.q Quad_Divide_Ceil(A.q, B.q)
    Protected Temp.q = A / B
    If (((a ! b) >= 0) And (a % b <> 0))
      ProcedureReturn Temp + 1
    Else
      ProcedureReturn Temp
    EndIf
  EndProcedure
  
  Procedure.d Color_Distance(Color_A.l, Color_B.l)
    ProcedureReturn Sqr(Pow(Red(Color_A) - Red(Color_B),2) + Pow(Green(Color_A) - Green(Color_B),2) + Pow(Blue(Color_A) - Blue(Color_B),2))
  EndProcedure
  
  Procedure.d Color_Distance_Squared(Color_A.l, Color_B.l)
    ProcedureReturn Pow(Red(Color_A) - Red(Color_B),2) + Pow(Green(Color_A) - Green(Color_B),2) + Pow(Blue(Color_A) - Blue(Color_B),2)
  EndProcedure
  
EndModule

; IDE Options = PureBasic 5.60 beta 6 (Windows - x64)
; CursorPosition = 34
; Folding = --
; EnableXP
; EnableCompileCount = 0
; EnableBuildCount = 0
; EnableExeConstant
; EnableUnicode