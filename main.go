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
	gray     = lipgloss.NewStyle().Foreground(lipgloss.Color("#696969"))
	green    = lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32"))
	darkgold = lipgloss.NewStyle().Foreground(lipgloss.Color("#B8860B"))
	gold     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	red      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	orange   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6347"))
	gameVar  game
)

type game struct {
	startTime time.Time // 紀錄遊玩時長
	start     bool      // 遊戲是否開始
	level     string    // 記錄難度
	intSize   int       // 元素個數,理應等於 l*l
	l         int       // 單行/列的元素個數
	bomb      int       // 炸彈總數
	msg       string    // 遊戲結束的訊息
	board     []string  // 遊戲真正的版面
}

// bubbletea使用的結構
type model struct {
	showBoard []string         // items on the to-do list
	cursor    int              // which to-do list item our cursor is pointing at
	selected  map[int]struct{} // which to-do items are selected
}

// 遊戲難度選單
func initialMenu() model {
	return model{
		showBoard: []string{"S", "M", "L"},
		selected:  make(map[int]struct{}),
	}
}

// 遊戲板
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
			if m.cursor-gameVar.l >= 0 {
				m.cursor -= gameVar.l
			}
		case "down":
			if m.cursor+gameVar.l < len(m.showBoard) {
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
			// 判別當前是"選單(false)"或是"遊戲(true)"
			if gameVar.start {
				if gameVar.board[m.cursor] == red.Render("M") {
					gameVar.msg = "踩到地雷了,遊戲結束\n"
					return m, tea.Quit
				} else if m.showBoard[m.cursor] == "*" || m.showBoard[m.cursor] == gray.Render("0") {
					checkBlank(m, m.cursor)
				}
				// 統計未點開的數量
				blankCount := 0
				for i := 0; i < gameVar.intSize; i++ {
					if m.showBoard[i] != "M" && m.showBoard[i] != "*" {
						blankCount += 1
					}
				}
				// 根據未點開的數量判別是否獲勝
				if blankCount == gameVar.intSize-gameVar.bomb {
					gameVar.msg = "找到所有炸彈了!\n"
					return m, tea.Quit
				}
			} else {
				// 難度選擇 S,M,L
				gameVar.level = m.showBoard[m.cursor]
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
		// 判別當前是"選單(false)"或是"遊戲(true)"
		if gameVar.start {
			if (i+1)%gameVar.l == 0 {
				s += fmt.Sprintf("%s%s\n", cursor, choice)

			} else {
				s += fmt.Sprintf("%s%s", cursor, choice)
			}
		} else {
			s += fmt.Sprintf("%s [%s] ", cursor, choice)
		}
	}
	// Send the UI for rendering
	return s
}
func main() {
	fmt.Println("S : 3 x 3 , 共 2 顆地雷\nM : 8 x 8 , 共 10 顆地雷\nL : 15 x 15 , 共 40 顆地雷 \n請選擇遊戲板大小:")
	menu := tea.NewProgram(initialMenu())
	if err := menu.Start(); err != nil {
		fmt.Printf("初始化menu失敗 : %v", err)
		os.Exit(1)
	}
	gameBoardInit()
	if !gameVar.start {
		fmt.Println("未選擇遊戲難度,遊戲結束")
		return
	}
	gameVar.startTime = time.Now()
	game := tea.NewProgram(initialModel())
	if err := game.Start(); err != nil {
		fmt.Printf("初始化遊戲板失敗 : %v", err)
		os.Exit(1)
	}
	if gameVar.msg != "" {
		s := " "
		for i := 0; i < gameVar.intSize; i++ {
			if (i+1)%gameVar.l == 0 {
				s += fmt.Sprintf("%s\n ", gameVar.board[i])
			} else {
				s += fmt.Sprintf("%s ", gameVar.board[i])
			}
		}
		fmt.Printf("%v%v耗時 : %v\n", gameVar.msg, s, time.Since(gameVar.startTime))
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
	// 根據難度設定參數
	switch gameVar.level {
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
		gameVar.start = false
		return
	}
	// 隨機產生炸彈的所在位置 => 隨機取不重複的數字bomb個
	var bombLocationSlice []int
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
	gameVar.board = make([]string, gameVar.intSize)
	for i := 0; i < gameVar.intSize; i++ {
		// M代表炸彈,B代表空白
		gameVar.board[i] = "B"
		if len(bombLocationSlice) > 0 && i == bombLocationSlice[0] {
			gameVar.board[i] = "M"
			bombLocationSlice = append(bombLocationSlice[:0], bombLocationSlice[1:]...)
		}
	}
	// 遍尋版面將旁邊有炸彈區塊填上數字
	// 每個區塊要檢查八個位子
	for i := 0; i < gameVar.intSize; i++ {
		var (
			l         = gameVar.l
			q         = i / l
			r         = i % l
			bombCount = 0
		)
		// 本身是炸彈就不檢查
		if gameVar.board[i] == "M" {
			gameVar.board[i] = red.Render(gameVar.board[i])
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
		// 填數字
		gameVar.board[i] = strconv.Itoa(bombCount)
		// 根據數字給予顏色
		switch bombCount {
		case 0:
			gameVar.board[i] = gray.Render(gameVar.board[i])
		case 1:
			gameVar.board[i] = green.Render(gameVar.board[i])
		case 2:
			gameVar.board[i] = darkgold.Render(gameVar.board[i])
		default:
			gameVar.board[i] = orange.Render(gameVar.board[i])
		}
	}
	gameVar.start = true
}

// 用於計算遊戲板每一格的周圍有多少炸彈
func checkBomb(i int) int {
	if i < 0 {
		return 0
	} else if i >= gameVar.intSize {
		return 0
	} else if gameVar.board[i] == "M" || gameVar.board[i] == red.Render("M") {
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
	} else if gameVar.board[i] == red.Render("M") {
		return
	} else if m.showBoard[i] != "*" {
		return
	} else if gameVar.board[i] == gray.Render("0") {
		m.showBoard[i] = gameVar.board[i]
		findBlank(m, i)
		return
	} else {
		m.showBoard[i] = gameVar.board[i]
		return
	}
}

func findBlank(m model, i int) {
	var (
		l = gameVar.l
		q = i / l
		r = i % l
	)
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
