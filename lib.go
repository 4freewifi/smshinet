package smshinet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"net"
	"time"
)

const (
	MsgTypeAuth          = 0  // 帳號密碼檢查
	MsgTypeSendText      = 1  // 傳送文字簡訊
	MsgTypeCheckText     = 2  // 查詢文字簡訊傳送結果
	MsgTypeRecvText      = 3  // 接收文字簡訊 (一般用戶不開放)
	MsgTypeSendWAPPush   = 13 // 傳送 WAP PUSH
	MsgTypeCheckWAPPush  = 14 // 查詢 WAP PUSH 傳送結果
	MsgTypeSendIntlText  = 15 // 傳送國際簡訊
	MsgTypeCheckIntlText = 2  // Same as MsgTypeCheckText
	MsgTypeCancelSchText = 16 // 取消預約文字簡訊
	MsgCodingBig5        = 1  // Big5
	MsgCodingBinary      = 2  // Binary
	MsgCodingUCS2        = 3  // unicode(UCS-2)
	MsgCodingUTF8        = 4  // unicode(UTF-8)
	MsgSetMaxLen         = 100
	MsgContentMaxLen     = 160
	RetSetMaxLen         = 80
	RetContentMaxLen     = 160
	MaxExpireMins        = 1440 // mmmm, 0001 ~ 1440 minutes
)

type SendMsg struct {
	// 訊息型態
	MsgType byte
	// 訊息編碼種類
	MsgCoding byte
	// 訊息優先權 (此功能不開放)
	MsgPriority byte
	// (保留用途)
	MsgCountryCode byte
	// 為msg_set[ ] 訊息內容的長度，注意：此處包含字串最後結尾的‘\0’符號
	MsgSetLen byte
	// 為msg_content[ ] 訊息內容的長度
	MsgContentLen byte
	// 訊息相關資料設定
	MsgSet [MsgSetMaxLen]byte
	// 簡訊內容
	MsgContent [MsgContentMaxLen]byte
}

type RecvMsg struct {
	RetCode       byte
	RetCoding     byte
	RetSetLen     byte
	RetContentLen byte
	RetSet        [RetSetMaxLen]byte
	RetContent    [RetContentMaxLen]byte
}

var AuthRetCode map[byte]error = map[byte]error{
	1: errors.New("Password error"),
	2: errors.New("The account not exist"),
	3: errors.New("Over the maximun allowed connection number"),
	4: errors.New("The account status not correct"),
	5: errors.New("get account data error"),
	6: errors.New("get password data error"),
	7: errors.New("System error, try again later"),
}

var SendRetCode map[byte]error = map[byte]error{
	1:  errors.New("Country code format error"),
	2:  errors.New("Coding format error"),
	3:  errors.New("Priority format error"),
	4:  errors.New("Msg_content_len format error"),
	5:  errors.New("Msg_content_len not the same with msg_content"),
	6:  errors.New("Telphone number format error"),
	7:  errors.New("Transfer type format error"),
	8:  errors.New("Limit time format error"),
	9:  errors.New("Ordered time format error"),
	10: errors.New("send to forign not allow now"),
	11: errors.New("Message sending failure, try again"),
	13: errors.New("wappush url length is zero"),
	14: errors.New("wappush msg_content length bigger than 88"),
	16: errors.New("message has 9-10 digits tel number"),
}

var CheckRetCode map[byte]error = map[byte]error{
	1:  errors.New("Mobile turn off/Mobile out of scope"),
	2:  errors.New("System contains no data"),
	3:  errors.New("MessageID format error"),
	4:  errors.New("has send to SMC, query no complete"),
	5:  errors.New("Ordered time beyond xx hours"),
	6:  errors.New("Send binary data to pager"),
	7:  errors.New("Code transfer fail"),
	8:  errors.New("telephone number or message content format error"),
	9:  errors.New("has expired at queue server"),
	10: errors.New("SMC without the data OR over re-transmission time"),
	15: errors.New("Message status unknown"),
	16: errors.New("Message sending failure"),
	17: errors.New("Message can not send to GSM/Pager"),
	18: errors.New("other error"),
	19: errors.New("Message is submitted to SMSC"),
	20: errors.New("reserve message, waiting send"),
	21: errors.New("reserve message, cancel send"),
	22: errors.New("message content deny"),
	23: errors.New("Message is barred by customer"),
}

var CommonRetCode map[byte]error = map[byte]error{
	30: errors.New("Message length is smaller than definition"),
	31: errors.New("network error, try again"),
	32: errors.New("msg_type not know"),
	40: errors.New("dataBase error"),
	41: errors.New("System internal error, try again later"),
	50: errors.New("ID/Password has not been checked"),
	51: errors.New("ID/Password checking again"),
	52: errors.New("text Service not apply yet"),
	53: errors.New("receive text service not apply yet"),
	58: errors.New("foreign message not apply yet"),
}

type Logger struct {
	Debugf   func(format string, args ...interface{})
	Infof    func(format string, args ...interface{})
	Errorf   func(format string, args ...interface{})
	Warningf func(format string, args ...interface{})
}

type Client struct {
	Addr string // something like 'api.hiair.hinet.net:8000'
	Dialer func(network, addr string) (net.Conn, error)
	Logger *Logger
	conn net.Conn
}

func (t *Client) dial(network, addr string) (net.Conn, error) {
	if t.Logger == nil {
		t.Logger = &Logger{
			Debugf: glog.V(1).Infof,
			Infof: glog.Infof,
			Errorf: glog.Errorf,
			Warningf: glog.Warningf,
		}
	}
	if t.Dialer != nil {
		c, err := t.Dialer(network, addr)
		if c == nil && err == nil {
			err = errors.New("smshinet.Dialer hook returned (nil, nil)")
		}
		return c, err
	}
	return net.Dial(network, addr)
}

func (t *Client) Dial() error {
	conn, err := t.dial("tcp", t.Addr)
	if err != nil {
		return err
	}
	t.conn = conn
	t.Logger.Infof("Connected to %s", t.Addr)
	return nil
}

func (t *Client) Close() error {
	return t.conn.Close()
}

func (t *Client) handleRecvMsg(ret *RecvMsg, definition map[byte]error) error {
	t.Logger.Debugf("RecvMsg %v", ret)
	if ret.RetContentLen > 0 {
		t.Logger.Debugf("RecvMsg.RetContent: %s",
			ret.RetContent[:ret.RetContentLen])
	}
	code := ret.RetCode
	if code == 0 {
		return nil
	}
	err, ok := definition[code]
	if ok {
		goto EXIT
	}
	err, ok = CommonRetCode[code]
	if ok {
		goto EXIT
	}
	err = fmt.Errorf("Unknown ret_code %d", code)
	t.Logger.Errorf("%s", err.Error())
	return err
EXIT:
	t.Logger.Warningf("handleRecvMsg: %s", err.Error())
	return err
}

func (t *Client) fillBytes(buf []byte, src string) (byte, error) {
	expect := len(src)
	n := copy(buf, src)
	if n != expect {
		return byte(n), fmt.Errorf("Short write: write %d expect %d", n, expect)
	}
	return byte(n), nil
}

func (t *Client) Auth(username, password string) error {
	l1 := len(username)
	if l1 > 8 {
		return errors.New("username too long, max 8")
	}
	l2 := len(password)
	if l2 > 8 {
		return errors.New("password too long, max 8")
	}
	msg := SendMsg{
		MsgType:   MsgTypeAuth,
		MsgCoding: MsgCodingBig5,
	}
	n, err := t.fillBytes(msg.MsgSet[:], username+"\x00"+password+"\x00")
	if err != nil {
		return err
	}
	msg.MsgSetLen = n
	t.Logger.Debugf("SendMsg %v", msg)
	err = binary.Write(t.conn, binary.BigEndian, msg)
	if err != nil {
		return err
	}
	ret := RecvMsg{}
	err = binary.Read(t.conn, binary.BigEndian, &ret)
	if err != nil {
		return err
	}
	err = t.handleRecvMsg(&ret, AuthRetCode)
	if err != nil {
		return err
	}
	t.Logger.Infof("%s Authenticated", username)
	return nil
}

func (t *Client) DialAndAuth(username, password string) error {
	err := t.Dial()
	if err != nil {
		return err
	}
	return t.Auth(username, password)
}

func (t *Client) sendTextNow(msg *SendMsg, recipient, message string,
	duration time.Duration) (msgId string, err error) {
	t.Logger.Infof("sendTextNow to %s expire %d: %s",
		recipient, int64(duration), message)
	l := len(message)
	if l > MsgContentMaxLen-1 {
		return "", fmt.Errorf("message too long, max %d",
			MsgContentMaxLen-1)
	}
	var n byte
	if duration == 0 {
		n, err = t.fillBytes(msg.MsgSet[:], recipient+"\x0001\x00")
		if err != nil {
			return
		}
	} else {
		mins := duration / time.Minute
		if mins > MaxExpireMins {
			err = fmt.Errorf(
				"Cannot set expire time longer then %d minutes",
				MaxExpireMins)
			return
		}
		s := fmt.Sprintf("%s\x0002\x00%4d\x00", recipient, mins)
		n, err = t.fillBytes(msg.MsgSet[:], s)
		if err != nil {
			return
		}
	}
	msg.MsgSetLen = n
	n, err = t.fillBytes(msg.MsgContent[:], message+"\x00")
	if err != nil {
		return
	}
	msg.MsgContentLen = n - 1
	t.Logger.Debugf("SendMsg %v", msg)
	err = binary.Write(t.conn, binary.BigEndian, msg)
	if err != nil {
		return
	}
	ret := RecvMsg{}
	err = binary.Read(t.conn, binary.BigEndian, &ret)
	if err != nil {
		return
	}
	err = t.handleRecvMsg(&ret, SendRetCode)
	if err != nil {
		return
	}
	if ret.RetContentLen == 0 {
		err = errors.New("Unexpected ret_content_len 0")
		return
	}
	msgId = string(ret.RetContent[:ret.RetContentLen])
	t.Logger.Infof("sendTextNow to %s succeeded with msgId %s",
		recipient, msgId)
	return
}

func (t *Client) SendTextInUTF8NowWithExpire(recipient, message string,
	duration time.Duration) (string, error) {
	l := len(recipient)
	if l > 10 {
		return "", errors.New("recipient number too long, max 10")
	}
	msg := SendMsg{
		MsgType:   MsgTypeSendText,
		MsgCoding: MsgCodingUTF8,
	}
	return t.sendTextNow(&msg, recipient, message, duration)
}

func (t *Client) SendTextInUTF8Now(recipient, message string) (
	string, error) {
	return t.SendTextInUTF8NowWithExpire(recipient, message, 0)
}

func (t *Client) SendIntlTextInUTF8NowWithExpire(recipient, message string,
	duration time.Duration) (
	string, error) {
	l := len(recipient)
	if l > 20 {
		return "", errors.New("recipient number too long, max 20")
	}
	msg := SendMsg{
		MsgType:   MsgTypeSendIntlText,
		MsgCoding: MsgCodingUTF8,
	}
	return t.sendTextNow(&msg, recipient, message, duration)
}

func (t *Client) SendIntlTextInUTF8Now(recipient, message string) (
	string, error) {
	return t.SendIntlTextInUTF8NowWithExpire(recipient, message, 0)
}

func (t *Client) CheckTextStatus(msgId string) error {
	msg := SendMsg{
		MsgType:   MsgTypeCheckText,
		MsgCoding: MsgCodingBig5,
	}
	n, err := t.fillBytes(msg.MsgSet[:], msgId+"\x00")
	if err != nil {
		return err
	}
	msg.MsgSetLen = n
	t.Logger.Debugf("SendMsg %v", msg)
	err = binary.Write(t.conn, binary.BigEndian, msg)
	if err != nil {
		return err
	}
	ret := RecvMsg{}
	err = binary.Read(t.conn, binary.BigEndian, &ret)
	if err != nil {
		return err
	}
	return t.handleRecvMsg(&ret, CheckRetCode)
}
