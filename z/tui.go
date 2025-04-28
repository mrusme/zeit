package z

import (
  "fmt"
  "math"
  "github.com/gookit/color"
  "github.com/shopspring/decimal"
)

func GetOutputBoxForNumber(number int, clr (func(...interface {}) string) ) (string) {
  switch(number) {
  case 0: return clr("  ")
  case 1: return clr(" ▄")
  case 2: return clr("▄▄")
  case 3: return clr("▄█")
  case 4: return clr("██")
  }

  return clr("  ")
}

func GetOutputBarForHours(hours decimal.Decimal, stats []Statistic) ([]string) {
  var bar = []string{
    color.FgGray.Render("····"),
    color.FgGray.Render("····"),
    color.FgGray.Render("····"),
    color.FgGray.Render("····"),
    color.FgGray.Render("····"),
    color.FgGray.Render("····"),
  }

  hoursInt := int((hours.Round(0)).IntPart())
  rest := ((hours.Round(0)).Mod(decimal.NewFromInt(4))).Round(0)
  restInt := int(rest.IntPart())

  divisible := hoursInt - restInt
  fullparts := divisible / 4

  colorsFull := make(map[int](func(...interface {}) string))
  colorsFullIdx := 0

  colorFraction := color.FgWhite.Render
  colorFractionPrevAmount := 0.0

  for _, stat := range stats {
    statHoursInt, _ := stat.Hours.Float64()
    statRest := (stat.Hours.Round(0)).Mod(decimal.NewFromInt(4))
    statRestFloat, _ := statRest.Float64()

    if statRestFloat > colorFractionPrevAmount {
      colorFractionPrevAmount = statRestFloat
      colorFraction = stat.Color
    }

    fullColoredParts := int(math.Round(statHoursInt) / 4)

    if fullColoredParts == 0 && statHoursInt > colorFractionPrevAmount {
      colorFractionPrevAmount = statHoursInt
      colorFraction = stat.Color
    }

    for i := 0; i < fullColoredParts; i++ {
      colorsFull[colorsFullIdx] = stat.Color
      colorsFullIdx++
    }
  }

  iColor := 0
  for i := (len(bar) - 1); i > (len(bar) - 1 - fullparts); i-- {
    if iColor < colorsFullIdx {
      bar[i] = " " + GetOutputBoxForNumber(4, colorsFull[iColor]) + " "
      iColor++
    } else {
      bar[i] = " " + GetOutputBoxForNumber(4, colorFraction) + " "
    }
  }

  if(restInt > 0) {
    bar[(len(bar) - 1 - fullparts)] = " " + GetOutputBoxForNumber(restInt, colorFraction) + " "
  }

  return bar
}

func OutputAppendRight(leftStr string, rightStr string, pad int) (string) {
  var output string = ""
  var rpos int = 0

  left := []rune(leftStr)
  leftLen := len(left)
  right := []rune(rightStr)
  rightLen := len(right)

  for lpos := 0; lpos < leftLen; lpos++ {
    if left[lpos] == '\n' || lpos == (leftLen - 1) {
      output = fmt.Sprintf("%s%*s", output, pad, "")
      for rpos = rpos; rpos < rightLen; rpos++ {
        output = fmt.Sprintf("%s%c", output, right[rpos])
        if right[rpos] == '\n' {
          rpos++
          break
        }
      }
      continue
    }
    output = fmt.Sprintf("%s%c", output, left[lpos])
  }

  return output
}

func GetColorFnFromHex(colorHex string) (func(...interface {}) string) {
  if colorHex == "" {
    colorHex = "#dddddd"
  }
  return color.NewRGBStyle(color.HEX(colorHex)).Sprint
}
