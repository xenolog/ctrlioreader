# Controlled io.Reader to test purposes

In this examples we immitate the network delay while bufio.Scan(...) works

```golang
import (
  "bufio"
  "bytes"
  "context"
  "testing"
  "testing/iotest"
  "time"
  assert "github.com/stretchr/testify/require"
  "github.com/xenolog/ctrlioreader"
)

func Test__ScanWithNetworkDelay(t *testing.T) {
  tt := assert.New(t)

  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

  part1 := `{"a":1,"b":2,"c":[1,2,3],`
  part2 := `"d":{"d1":"aaa","d2":"bbb"}}`
  expected := part1 + part2

  dataStream := bytes.NewBuffer(make([]byte, 0, 256))

  rdr, eof := ctrlioreader.NewCtrlReader(ctx, dataStream, 0)
  scan := bufio.NewScanner(iotest.NewReadLogger(">>>", rdr))

  scan.Split(ScanJSONobject) //    <---- tested function

  dataStream.WriteString(part1)
  go func() {
    time.Sleep(3 * time.Second)
    dataStream.WriteString(part2)
    time.Sleep(2 * time.Second)
    eof()
  }()

  tt.True(scan.Scan())
  tt.NoError(scan.Err())
  tt.EqualValues(expected, scan.Text())
}
```
