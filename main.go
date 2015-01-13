// Copyright (c) 2014  by f4lk0r_ 
// Licensed under the GPLv3, see the License.txt file for details

package main

import (
    //janimo's textsecure library. Documentation here: https://godoc.org/github.com/janimo/textsecure
	ts "github.com/janimo/textsecure"
    //go ncurses library. Documentation here: https://godoc.org/code.google.com/p/goncurses
    gc "code.google.com/p/goncurses"

    "log"
)

//var telToName map[string]string
var debugLog = startDebugLogger()
var inputBuffer []byte

//sets up the initial configuration of curses. Keeps code in main cleaner.
func configCurses(stdscr *gc.Window) {
    if !gc.HasColors() {
        log.Fatal("Example requires a colour capable terminal")
    }
    if err := gc.StartColor(); err != nil {
        log.Fatal("Starting Colors failed:",err)
    }
    gc.Echo(false)

    if err := gc.InitPair(1, gc.C_RED, gc.C_BLACK); err != nil {
        log.Fatal("InitPair failed: ", err)
    }
    gc.InitPair(2, gc.C_BLUE, gc.C_BLACK)
    gc.InitPair(3, gc.C_GREEN, gc.C_BLACK)
    gc.InitPair(4, gc.C_YELLOW, gc.C_BLACK)
    gc.InitPair(5, gc.C_CYAN, gc.C_BLACK)
    gc.InitPair(6, gc.C_MAGENTA, gc.C_WHITE)
    gc.InitPair(7, gc.C_MAGENTA, gc.C_BLACK)

    //set background color to black
    stdscr.SetBackground(gc.Char(' ') | gc.ColorPair(0))

    stdscr.Keypad(true)
    gc.Cursor(1)
    gc.CBreak(true)
    stdscr.Clear()
    stdscr.ColorOn(1)
}

//creates the three Main big windows that make up the GUI
func createMainWindows(stdscr *gc.Window) (*gc.Window, *gc.Window, *gc.Window, *gc.Window) {
    rows, cols := stdscr.MaxYX()
    height, width := rows,int(float64(cols) * .2)

    var contactsWin,messageWin,inputBorderWin, inputWin *gc.Window
    var err error
    contactsWin, err = gc.NewWindow(height, width, 0, 0)
    if err != nil {
        log.Fatal("Failed to create Contact Window:",err)
    }
    err = contactsWin.Border('|','|','-','-','+','+','+','+')
    if err != nil {
        log.Fatal("Failed to create Border of Contact Window:",err)
    }

    begin_x := width+1
    height = int(float64(rows) * .8)
    width =  int(float64(cols) * .8)
    messageWin, err = gc.NewWindow(height, width, 0, begin_x)
    if err != nil {
        log.Fatal("Failed to create Message Window:",err)
    }
    err = messageWin.Border('|','|','-','-','+','+','+','+')
    if err != nil {
        log.Fatal("Failed to create Border of Message Window:",err)
    }

    begin_y := int(float64(rows) * .8 ) + 1
    height = int(float64(rows) * .2)
    inputBorderWin, err = gc.NewWindow(height, width, begin_y, begin_x)
    if err != nil {
        log.Fatal("Failed to create InputBorder Window:",err)
    }
    err = inputBorderWin.Border('|','|','-','-','+','+','+','+')
    if err != nil {
        log.Fatal("Failed to create Border of the InputBorder Window:",err)
    }
    // inputBorderWin is just the border. InputWin is the actual input window
    // doing this simplifies handling text a fair amount.
    // also the derived function never errors. Which seems dangerous 
    inputWin, err = gc.NewWindow((height -2),(width-2),(begin_y+1),(begin_x+1))
    if err != nil {
        log.Fatal("Failed to create the inputWin Window:",err)
    }

    inputWin.Keypad(true)

    return contactsWin,messageWin,inputBorderWin,inputWin
}

func messageHandler(msg *ts.Message) {
    err := ts.SendMessage(msg.Source(), msg.Message())
    if err != nil {
        log.Println(err)
    }
    return
}

// handles keyboard input
func inputHandler( inputWin *gc.Window, stdscr *gc.Window) {
    var c gc.Char
    var rawInput gc.Key
    max_y, max_x := inputWin.MaxYX()
    for {
        rawInput = inputWin.GetChar()
        c = gc.Char(rawInput)
        //debugLog.Println(rawInput)
        //debugLog.Println(c)

        //Escape to Quit
        if c == gc.Char(27) {
            break
        } else if rawInput == gc.KEY_BACKSPACE {
            //Delete Key
            y,x := inputWin.CursorYX()
            if x != 0 {
                inputWin.MoveDelChar(y,x-1)
                inputBuffer = inputBuffer[0:len(inputBuffer)-1]
                //debugLog.Println(inputBuffer)
            } else {
                inputWin.MoveDelChar(y-1,max_x)
                inputBuffer = inputBuffer[0:len(inputBuffer)-1]
                //debugLog.Println(inputBuffer)
            }
        } else if c == gc.KEY_LEFT {
            y,x := inputWin.CursorYX()
            if x != 0 {
                inputWin.Move(y,x-1)
            }
        } else if c == gc.KEY_RIGHT {
            y,x := inputWin.CursorYX()
            if x != max_x {
                inputWin.Move(y,x+1)
            }
        } else if c == gc.KEY_UP {
            y,x := inputWin.CursorYX()
            if y != 0 {
                inputWin.Move(y-1,x)
            }
        } else if c == gc.KEY_DOWN {
            y,x := inputWin.CursorYX()
            if y != max_y {
                inputWin.Move(y+1,x)
            }
        } else if rawInput == gc.KEY_RETURN {
            clearScrSendMsg(inputWin)
        } else if rawInput == gc.KEY_SRIGHT {
            inputWin.Print("\n")
            inputBuffer = append(inputBuffer,byte(10))
        } else if rawInput == gc.KEY_SLEFT {
        } else {
            inputWin.Print(string(c))
            inputBuffer = append(inputBuffer,byte(c))
            //debugLog.Println(inputBuffer)
        }
    }

}

// In addition to sending a message using janimo's library, also clears screen and resets buffer
func clearScrSendMsg (inputWin *gc.Window) {
    if len(inputBuffer) != 0 {
        msg := string(inputBuffer)
        to := string("+12345678910")
        err := ts.SendMessage(to,msg)
        if err != nil {
            gc.End()
            log.Fatal("SendMessage failed yo: ",err)
        }
        inputBuffer = []byte{}
        inputWin.Erase()
    }
}

// Hello dialog. So the nooblets know the controls
func doHello( stdscr *gc.Window) {
    stdscr.Refresh()
    stdscr.MovePrintln(5,5, "Controls:")
    stdscr.Println("Escape: Exits the program.")
    stdscr.Println("Tab: Switches between the input window and the Message window.")
    stdscr.Println("Return: Sends a message")
    stdscr.Println(`Shift + Right Arrow: puts in a new line '\n' character (like shift+return in facebook chat)`)
    stdscr.Println("Press any key to continue...")
    stdscr.GetChar()
    stdscr.Clear()
    stdscr.Refresh()
}


// creates a curses based TUI for the textsecure library
func main() {
    stdscr, err := gc.Init()
    client := &ts.Client{
        RootDir:        ".",
        ReadLine:       ts.ConsoleReadLine,
        MessageHandler: messageHandler,
    }
    ts.Setup(client)
    if err != nil {
        log.Fatal("Error initializing curses:", err)
    }
    defer gc.End()
    configCurses(stdscr)
    doHello(stdscr)

    contacts, err := ts.GetRegisteredContacts()
    if err != nil {
        log.Fatal("Could not get contacts: %s\n", err)
    }

    menu_items := make([]*gc.MenuItem, len(contacts))
    for i, val := range contacts {
        menu_items[i], err = gc.NewItem((val.Name), "")
        if err != nil {
            log.Fatal("Error making item for contact menu... ", err)
        }
        defer menu_items[i].Free()
    }
    contactMenu, err := gc.NewMenu(menu_items)
    if err != nil {
        log.Fatal("Error making contact menu... ", err)
    }
    defer contactMenu.Free()

    contactsWin, messageWin, inputBorderWin, inputWin := createMainWindows(stdscr)

    contactsWinSizeY, contactsWinSizeX := contactsWin.MaxYX()
    contactsWin.Keypad(true)
    contactsMenuWin := contactsWin.Derived((contactsWinSizeY-5),(contactsWinSizeX-2),3,1)
    contactMenu.SetWindow(contactsMenuWin)
    contactMenu.Format(5,1)
    contactMenu.Mark(" * ")

    title := "Contacts"
    contactsWin.MovePrint(1, (contactsWinSizeX/2)-(len(title)/2), title)
    contactsWin.HLine(2, 1, '-', contactsWinSizeX-2)

    contactMenu.Post()
    defer contactMenu.UnPost()
    contactsWin.Refresh()

    messageWin.Refresh()
    inputBorderWin.Refresh()

    inputWin.Move(0,0)
    inputHandler(inputWin, stdscr)
}
