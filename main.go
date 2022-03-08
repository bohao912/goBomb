package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 顏色
var (
	cyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32"))
	gold   = lipgloss.NewStyle().Foreground(lipgloss.Color("#B8860B"))
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	orange = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6347"))
)
var gameVar game

type game struct {
	start     time.Time // 紀錄遊玩時長
	intSize   int       // 元素個數,理應等於 l*l
	l         int       // 單行/列的元素個數
	bomb      int       // 炸彈總數
	gameMsg   string    // 遊戲結束的訊息
	gameBoard []string  // 遊戲真正的版面
}

// bubbletea使用的結構
type model struct {
	showBoard []string         // items on the to-do list
	cursor    int              // which to-do list item our cursor is pointing at
	selected  map[int]struct{} // which to-do items are selected
}

func initialModel() model {
	return model{
		showBoard: showBoardInit(),
		selected:  make(map[int]struct{}),
	}
}
func (m model) Init() tea.Cmd {
	return nil
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// 偵測按鍵按下
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		// 指針移動
		case "up":
			if m.cursor-4 >= 0 {
				m.cursor -= gameVar.l
			}
		case "down":
			if m.cursor+4 <= len(m.showBoard) {
				m.cursor += gameVar.l
			}
		case "right":
			if m.cursor < len(m.showBoard)-1 {
				m.cursor++
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		// click 點擊
		case "enter":
			if gameVar.gameBoard[m.cursor] == red.Render("M") {
				gameVar.gameMsg = "踩到地雷了,遊戲結束\n"
				return m, tea.Quit
			} else if m.showBoard[m.cursor] == "*" || m.showBoard[m.cursor] == cyan.Render("0") {
				checkBlank(m, m.cursor)
			}
			blankCount := 0
			for i := 0; i < gameVar.intSize; i++ {
				if m.showBoard[i] != "M" && m.showBoard[i] != "*" {
					blankCount += 1
				}
			}
			if blankCount == gameVar.intSize-gameVar.bomb {
				gameVar.gameMsg = "找到所有炸彈了!\n"
				return m, tea.Quit
			}
		}
	}
	return m, nil
}
func (m model) View() string {
	s := ""
	for i, choice := range m.showBoard {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		if (i+1)%gameVar.l == 0 {
			s += fmt.Sprintf("%s [%s]\n\n", cursor, choice)

		} else {
			s += fmt.Sprintf("%s [%s] ", cursor, choice)
		}
	}
	// Send the UI for rendering
	return s
}
func main() {
	fmt.Println("S : 3 x 3 , 共 2 顆地雷\nM : 8 x 8 , 共 10 顆地雷\nL : 15 x 15 , 共 40 顆地雷 \n請選擇遊戲板大小:")
	gameBoardInit()
	gameVar.start = time.Now()
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	if gameVar.gameMsg != "" {
		s := ""
		for i := 0; i < gameVar.intSize; i++ {
			if (i+1)%gameVar.l == 0 {
				s += fmt.Sprintf(" [%s]\n", gameVar.gameBoard[i])
			} else {
				s += fmt.Sprintf(" [%s] ", gameVar.gameBoard[i])
			}
		}
		fmt.Printf("%v%v耗時 : %v\n", gameVar.gameMsg, s, time.Since(gameVar.start))
	}
}
func showBoardInit() []string {
	showBoard := make([]string, gameVar.intSize)
	for i := 0; i < gameVar.intSize; i++ {
		showBoard[i] = "*"
	}
	return showBoard
}

// 初始化遊戲板
func gameBoardInit() {
	strSize := ""
	fmt.Scanln(&strSize)
	switch strSize {
	case "S":
		gameVar.intSize = 9
		gameVar.l = 3
		gameVar.bomb = 2
	case "M":
		gameVar.intSize = 64
		gameVar.l = 8
		gameVar.bomb = 10
	case "L":
		gameVar.intSize = 255
		gameVar.l = 15
		gameVar.bomb = 40
	default:
		fmt.Println("請輸入S,M,L其中一個大小")
		gameBoardInit()
		return
	}
	// 隨機產生炸彈的所在位置 => 隨機取不重複的數字bomb個
	bombLocationSlice := make([]int, 0)
	rand.Seed(time.Now().UnixNano())
	for len(bombLocationSlice) < gameVar.bomb {
		bombLocation := rand.Intn(gameVar.intSize)
		exist := false
		for _, v := range bombLocationSlice {
			if v == bombLocation {
				exist = true
				break
			}
		}
		if !exist {
			bombLocationSlice = append(bombLocationSlice, bombLocation)
		}
	}
	sort.Ints(bombLocationSlice)
	// 遊戲真正的版面
	gameVar.gameBoard = make([]string, gameVar.intSize)
	for i := 0; i < gameVar.intSize; i++ {
		// M代表炸彈,B代表空白
		if len(bombLocationSlice) > 0 {
			if i == bombLocationSlice[0] {
				gameVar.gameBoard[i] = "M"
				bombLocationSlice = append(bombLocationSlice[:0], bombLocationSlice[1:]...)
			} else {
				gameVar.gameBoard[i] = "B"
			}
		} else {
			gameVar.gameBoard[i] = "B"
		}
	}
	// 遍尋版面將旁邊有炸彈區塊填上數字
	// 每個區塊要檢查八個位子
	for i := 0; i < gameVar.intSize; i++ {
		l := gameVar.l
		q := i / l
		r := i % l
		bombCount := 0
		if gameVar.gameBoard[i] == "M" {
			gameVar.gameBoard[i] = red.Render(gameVar.gameBoard[i])
			continue
		}
		if (i-l-1)/l == q-1 && (i-l-1)%l == r-1 {
			bombCount += checkBomb(i - l - 1)
		}
		if (i-l)/l == q-1 && (i-l)%l == r {
			bombCount += checkBomb(i - l)
		}
		if (i-l+1)/l == q-1 && (i-l+1)%l == r+1 {
			bombCount += checkBomb(i - l + 1)
		}
		if (i-1)/l == q && (i-1)%l == r-1 {
			bombCount += checkBomb(i - 1)
		}
		if (i+1)/l == q && (i+1)%l == r+1 {
			bombCount += checkBomb(i + 1)
		}
		if (i+l-1)/l == q+1 && (i+l-1)%l == r-1 {
			bombCount += checkBomb(i + l - 1)
		}
		if (i+l)/l == q+1 && (i+l)%l == r {
			bombCount += checkBomb(i + l)
		}
		if (i+l+1)/l == q+1 && (i+l+1)%l == r+1 {
			bombCount += checkBomb(i + l + 1)
		}
		gameVar.gameBoard[i] = strconv.Itoa(bombCount)
		switch bombCount {
		case 0:
			gameVar.gameBoard[i] = cyan.Render(gameVar.gameBoard[i])
		case 1:
			gameVar.gameBoard[i] = green.Render(gameVar.gameBoard[i])
		case 2:
			gameVar.gameBoard[i] = gold.Render(gameVar.gameBoard[i])
		default:
			gameVar.gameBoard[i] = orange.Render(gameVar.gameBoard[i])
		}
	}
}

// 用於計算遊戲板每一格的周圍有多少炸彈
func checkBomb(i int) int {
	if i < 0 {
		return 0
	} else if i >= gameVar.intSize {
		return 0
	} else if gameVar.gameBoard[i] == "M" || gameVar.gameBoard[i] == red.Render("M") {
		return 1
	} else {
		return 0
	}
}

// 用於揭露點選的附近八格,地雷不揭露,若為0繼續則繼續揭露,有數字只揭露自己
func checkBlank(m model, i int) {
	if i < 0 {
		return
	} else if i >= gameVar.intSize {
		return
	} else if gameVar.gameBoard[i] == red.Render("M") {
		return
	} else if m.showBoard[i] != "*" {
		return
	} else if gameVar.gameBoard[i] == cyan.Render("0") {
		m.showBoard[i] = gameVar.gameBoard[i]
		findBlank(m, i)
		return
	} else {
		m.showBoard[i] = gameVar.gameBoard[i]
		return
	}
}

func findBlank(m model, i int) {
	l := gameVar.l
	q := i / l
	r := i % l
	if (i-l-1)/l == q-1 && (i-l-1)%l == r-1 {
		checkBlank(m, i-l-1)
	}
	if (i-l)/l == q-1 && (i-l)%l == r {
		checkBlank(m, i-l)
	}
	if (i-l+1)/l == q-1 && (i-l+1)%l == r+1 {
		checkBlank(m, i-l+1)
	}
	if (i-1)/l == q && (i-1)%l == r-1 {
		checkBlank(m, i-1)
	}
	if (i+1)/l == q && (i+1)%l == r+1 {
		checkBlank(m, i+1)
	}
	if (i+l-1)/l == q+1 && (i+l-1)%l == r-1 {
		checkBlank(m, i+l-1)
	}
	if (i+l)/l == q+1 && (i+l)%l == r {
		checkBlank(m, i+l)
	}
	if (i+l+1)/l == q+1 && (i+l+1)%l == r+1 {
		checkBlank(m, i+l+1)
	}
}
