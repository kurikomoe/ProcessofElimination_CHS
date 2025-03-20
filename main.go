package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"unicode"
	"unicode/utf8"
)

var mm map[string]string = make(map[string]string)

var text_out *os.File

func containsJapaneseAndPunctuation(b []byte) bool {
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r == utf8.RuneError {
			return false
		}
		if isJapanese(r) {
			return true
		}
		b = b[size:]
	}
	return false
}
func isJapanese(r rune) bool {
	return unicode.In(r, &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x3040, 0x309F, 1}, // 平假名
			{0x30A0, 0x30FF, 1}, // 片假名
			{0x4E00, 0x9FFF, 1}, // 汉字
			{0xFF65, 0xFF9F, 1}, // 半角片假名
		},
	})
}
func isPunctuation(r rune) bool {
	return unicode.In(r, &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x3000, 0x303F, 1}, // 常用中文标点
			{0x0020, 0x007E, 1}, // 常用英文标点
		},
	})
}
func isEnglish(r rune) bool {
	return unicode.In(r, &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x0041, 0x005A, 1}, // 大写字母
			{0x0061, 0x007A, 1}, // 小写字母
		},
	})
}


type EntryData struct {
  raw []byte
  rawSize int

  isStr []bool
  data  [][]byte  // with trailing zero for string
}

func (this *EntryData) findText() {
  // fmt.Println("raw")
  // fmt.Printf("%x\n", this.raw)

  buf := []byte {}

  for i := 0; i < this.rawSize; i++ {
    d := this.raw[i]
    if d == 0x10 { // start of string
      var size uint16;

      slice := this.raw[i+1:]
      r := bytes.NewReader(slice)

      err := binary.Read(r, binary.LittleEndian, &size)
      if err != nil {
        // fmt.Println(size)
        goto Exit
      }

      if size == 0 {
        goto Exit
      }

      str := make([]byte, size)
      err = binary.Read(r, binary.LittleEndian, &str)
      if err != nil {
        goto Exit
      }

      // fmt.Println(i)
      // fmt.Println(size)
      // fmt.Printf("%x\n", str)
      // fmt.Printf("%x\n", this.raw[i:uint16(i)+size+2])

      if str[size-1] != 0 {
        goto Exit
      }

      // fmt.Println(string(str))
      // bufio.NewReader(os.Stdin).ReadString('\n')

      flag := true
      for i := 0; i < int(size)-1; i++ {
        if str[i] == 0 {
          flag = false
          break
        }
      }

      if !flag {
        goto Exit
      }

      // Real string
      if len(buf) != 0 {
        this.isStr = append(this.isStr, false)
        this.data = append(this.data, buf)
        buf = []byte {}
      }

      // remove trailing zero
      str = str[0:size-1]

      text_out.Write(str)
      text_out.WriteString("\n")


      this.isStr = append(this.isStr, true)

      if v, ok := mm[string(str)]; ok {
        if v == "" { v = "测试文本" }
        this.data = append(this.data, []byte(v))
      } else {
        // this.data = append(this.data, str)
        this.data = append(this.data, []byte("测试文本"))
      }

      i += int(size) + 0x2
      continue
    }

Exit:
    buf = append(buf, d)
  }

  if len(buf) != 0 {
    this.isStr = append(this.isStr, false)
    this.data = append(this.data, buf)
    buf = []byte {}
  }
}

func (this *EntryData)Write(file io.Writer) {
  // binary.Write(file, binary.LittleEndian, this.raw)
  for i := 0; i < len(this.isStr); i++ {
    if !this.isStr[i] {
      binary.Write(file, binary.LittleEndian, this.data[i])
    } else {
      d := this.data[i]
      binary.Write(file, binary.LittleEndian, uint8(0x10))
      binary.Write(file, binary.LittleEndian, uint16(len(d)+1))
      binary.Write(file, binary.LittleEndian, d)
      binary.Write(file, binary.LittleEndian, uint8(0))
    }
  }
}


type DatEntry struct {
  _id    int
  size   int

  id     uint32
  dwUn0  uint32
  offset uint32
}

func (entry *DatEntry) Read(file *os.File) {
  binary.Read(file, binary.LittleEndian, &entry.id)
  binary.Read(file, binary.LittleEndian, &entry.dwUn0)
  binary.Read(file, binary.LittleEndian, &entry.offset)
}

func (entry *DatEntry) Write(file *os.File) {
  binary.Write(file, binary.LittleEndian, entry.id)
  binary.Write(file, binary.LittleEndian, entry.dwUn0)
  binary.Write(file, binary.LittleEndian, entry.offset)
}

type Dat struct {
  dwEntrySize   uint32
  dwEntryCount  uint32

  aEntries      []DatEntry

  aData      []EntryData
}

func (this *Dat)Write(file *os.File) {
  binary.Write(file, binary.LittleEndian, this.dwEntrySize)
  binary.Write(file, binary.LittleEndian, this.dwEntryCount)

  for _, entry := range(this.aEntries) {
    entry.Write(file)
  }

  for _, data := range(this.aData) {
    data.Write(file)
  }
}



func main() {
  fmt.Fprintln(os.Stderr, "Hello World!");

  text_out, _ = os.Create("texts.txt")
  defer text_out.Close()

  zhFile, err := os.Open("texts.ja.zh.txt")
  if err != nil {
    panic(err)
  }
  defer zhFile.Close()


  scanner := bufio.NewScanner(zhFile)
  lineNumber := 1
  var jaLine string
  for scanner.Scan() {
    line := scanner.Text() // 获取当前行内容
    if lineNumber%2 == 1 {
      // 奇数行：日文，作为 key
      jaLine = line
    } else {
      // 偶数行：中文，作为 value
      mm[jaLine] = line
    }
    lineNumber++
  }

  // ------------------------------------

  filePath := "Script/Talk.dat"
  file, err := os.Open(filePath)
  if err != nil {
    fmt.Fprintln(os.Stderr, filePath)
    return
  }
  defer file.Close()

  fileInfo, err := file.Stat()
  fileSize := fileInfo.Size()

  var dat Dat
  binary.Read(file, binary.LittleEndian, &dat.dwEntrySize)
  binary.Read(file, binary.LittleEndian, &dat.dwEntryCount)

  // fmt.Println(dat)

  if (int64)(dat.dwEntrySize * dat.dwEntryCount) > fileSize {
    panic("Invaid size")
  }

  dat.aEntries = make([]DatEntry, dat.dwEntryCount)

  // fmt.Println("--------------")
  for i := 0; i < int(dat.dwEntryCount); i++ {
    var entry DatEntry
    entry._id = i
    entry.Read(file)
    dat.aEntries[i] = entry
    // fmt.Printf("{{ %x, %x, %x }}\n", entry.id, entry.dwUn0, entry.offset)
  }
  // fmt.Println("--------------")

  dataBegPos, err := file.Seek(int64(0x8 + dat.dwEntryCount * dat.dwEntrySize), io.SeekStart)

  sort.Slice(dat.aEntries, func(i, j int) bool {
    return dat.aEntries[i].offset < dat.aEntries[j].offset
  })

  // Get Entry Size
  for i, entry := range(dat.aEntries) {
    if i + 1 == len(dat.aEntries) {
      // The last one
      dat.aEntries[i].size = int(fileSize) - int(dataBegPos) - int(entry.offset)
    } else {
      dat.aEntries[i].size = int(dat.aEntries[i+1].offset) - int(entry.offset)
    }
  }


  dat.aData = make([]EntryData, dat.dwEntryCount)

  fmt.Println("total count: ", len(dat.aEntries))

  for i, entry := range(dat.aEntries) {
    // fmt.Println(i, entry.offset, entry.size)
    curPos := dataBegPos + int64(entry.offset)
    file.Seek(curPos, io.SeekStart)

    dat.aData[i].rawSize = entry.size;
    dat.aData[i].raw = make([]byte, min(int64(entry.size)+0x20, fileSize-curPos))

    err := binary.Read(file, binary.LittleEndian, &dat.aData[i].raw)
    if err != nil {
      panic(err)
    }

    dat.aData[i].findText()

  }

  curPos := uint32(0);
  for i := 0; i < int(dat.dwEntryCount); i++ {
    dat.aEntries[i].offset = curPos;
    buffer := new(bytes.Buffer)
    w := bufio.NewWriter(buffer)
    dat.aData[i].Write(w)
    w.Flush()
    buf := buffer.Bytes()
    curPos += uint32(len(buf))
  }

  sort.Slice(dat.aEntries, func(i, j int) bool {
    return dat.aEntries[i]._id < dat.aEntries[j]._id
  })


  fout, err := os.Create("Talk.dat")
  if err != nil {
    panic(err)
  }
  dat.Write(fout)
  defer fout.Close()
}
