package z

import (
  "fmt"
)

const(
  TUI_ROWS = 200
  TUI_COLS = 200
)

type TuiBuffer [TUI_ROWS][TUI_COLS]rune

type Tui struct {
  Buffer TuiBuffer
}

func (tui *Tui) Init() {
  for row := 0; row < TUI_ROWS; row++ {
    for col := 0; col < TUI_COLS; col++ {
      tui.Buffer[row][col] = 0
    }
  }
}

func (tui *Tui) Merge(buffer TuiBuffer, x int, y int) (bool) {
  inputRow := 0
  inputCol := 0

  for row := x; row < TUI_ROWS; row++ {
    for col := y; col < TUI_COLS; col++ {
      if buffer[inputRow][inputCol] != 0 {
        tui.Buffer[row][col] = buffer[inputRow][inputCol]
      }
      inputCol++
    }
    inputRow++
    inputCol = 0
  }

  return true
}

func (tui *Tui) Render(cols int, rows int) (string) {
  var output string = ""
  var emptyRow bool = true

  for row := 0; row < rows; row++ {
    for col := 0; col < cols; col++ {
      var chr rune = ' '

      if tui.Buffer[row][col] != 0 {
        chr = tui.Buffer[row][col]
        emptyRow = false
      } else {
        chr = ' '
      }
      output = fmt.Sprintf("%s%c", output, chr)
    }
    if emptyRow == false {
      output = fmt.Sprintf("%s\n", output)
      emptyRow = true
    }
  }

  return output
}
