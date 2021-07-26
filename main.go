package teamones_helper

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type QueryParam struct {
	Fields string                 `json:"fields"`
	Filter map[string]interface{} `json:"filter"`
	Page   []interface{}          `json:"page"`
	Order  string                 `json:"order"`
}

type SelectQueryParam struct {
	Fields string                 `json:"fields"`
	Limit  float64                `json:"limit"`
	Offset float64                `json:"offset"`
	Order  string                 `json:"order"`
	Where  map[string]interface{} `json:"where"`
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// 生成md5值
func GenerateMd5(str string) string {
	w := md5.New()
	io.WriteString(w, str)
	md5str := fmt.Sprintf("%x", w.Sum(nil))
	return md5str
}

// 获取文件md5值
func GetFileMd5(filePath string) (MD5Str string) {
	md5 := md5.New()

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("read fail", err)
	}

	io.Copy(md5, file)
	MD5Str = hex.EncodeToString(md5.Sum(nil))
	file.Close()
	return
}

// 解析值
func ParseSingleVal(v interface{}) interface{} {
	valType := reflect.TypeOf(v)
	typeString := valType.String()
	if typeString == "json.Number" {
		n, _ := v.(json.Number)
		if i, err := n.Int64(); err == nil {
			return i
		} else if f, err := n.Float64(); err == nil {
			return f
		} else {
			return v
		}
	} else {
		return StrVal(v)
	}
}

// 生成select查询过滤SQL
func GenerateSelectQueryParam(Param QueryParam, TableName string) (SqlParam SelectQueryParam) {
	sqlFields := "*"

	// 查询字段
	if Param.Fields != "" {
		sqlFields = Param.Fields
	}
	SqlParam.Fields = sqlFields

	// 分页
	if Param.Page != nil && len(Param.Page) > 0 {
		page := (Param.Page[0]).(float64)
		pageSize := (Param.Page[1]).(float64)
		SqlParam.Limit = pageSize
		SqlParam.Offset = (page - 1) * pageSize
	} else {
		// 最大1000条
		SqlParam.Limit = viper.GetFloat64("mysql.maxQueryNumber")
		if !(SqlParam.Limit > 0) {
			SqlParam.Limit = 1000
		}

		SqlParam.Offset = 0
	}

	// 排序
	if Param.Order != "" {
		OrderArray := strings.Split(Param.Order, ",")
		SqlParam.Order = strings.Join(OrderArray, " ")
	}

	// 处理过滤条件
	if Param.Filter != nil && len(Param.Filter) > 0 {
		whereParam := map[string]interface{}{}
		for field, val := range Param.Filter {
			valType := reflect.TypeOf(val)
			typeString := valType.String()
			if typeString == "json.Number" {
				n, _ := val.(json.Number)
				if i, err := n.Int64(); err == nil {
					whereParam["eq_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
						"condition": field + " = ?",
						"val":       i,
					}
				} else if f, err := n.Float64(); err == nil {
					whereParam["eq_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
						"condition": field + " = ?",
						"val":       f,
					}
				}
			} else if InArray(valType.String(), []interface{}{"string", "float64"}) {
				whereParam["eq_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
					"condition": field + " = ?",
					"val":       val,
				}
			} else if valType.String() == "[]interface {}" {
				if val.([]interface{})[0] != nil {
					condition := val.([]interface{})[0]
					vals := ParseSingleVal(val.([]interface{})[1])

					switch condition {
					case "-eq":
						whereParam["eq_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " = ?",
							"val":       vals,
						}
						break
					case "-neq":
						whereParam["neq_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " <> ?",
							"val":       vals,
						}
						break
					case "-lk":
						whereParam["like_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " LIKE ?",
							"val":       "%" + vals.(string) + "%",
						}
						break
					case "-not-lk":
						whereParam["not_like_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " NOT LIKE ?",
							"val":       "%" + vals.(string) + "%",
						}
						break
					case "-gt":
						whereParam["gt_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " > ?",
							"val":       vals,
						}
						break
					case "-egt":
						whereParam["egt_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " >= ?",
							"val":       vals,
						}
						break
					case "-lt":
						whereParam["lt_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " < ?",
							"val":       vals,
						}
						break
					case "-elt":
						whereParam["elt_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " <= ?",
							"val":       vals,
						}
						break
					case "-in":
						whereParam["in_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " in (?)",
							"val":       strings.Split(vals.(string), ","),
						}
						break
					case "-not-in":
						whereParam["not_in_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " not in (?)",
							"val":       strings.Split(vals.(string), ","),
						}
						break
					case "-bw":
						bwVal := strings.Split(vals.(string), ",")
						whereParam["bw_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " BETWEEN ? AND ?",
							"val1":      bwVal[0],
							"val2":      bwVal[1],
						}
						break
					case "-not-bw":
						bwNoteVal := strings.Split(vals.(string), ",")
						whereParam["bw_"+RandStringBytesMaskImprSrcUnsafe(6)] = map[string]interface{}{
							"condition": field + " NOT BETWEEN ? AND ?",
							"val1":      bwNoteVal[0],
							"val2":      bwNoteVal[1],
						}
						break
					}
				}
			}
		}
		SqlParam.Where = whereParam
	}

	return
}

// 判断是否在数字内
func InArray(need interface{}, needArr []interface{}) bool {
	for _, v := range needArr {
		if need == v {
			return true
		}
	}
	return false
}

// 随机字符串字节掩码
func RandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

//获取文件的扩展名
func GetFileExt(path string) string {
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			return strings.ToLower(path[(i + 1):])
		}
	}
	return ""
}

// 时间区间转换成秒为单位
func DurToSec(dur string) (sec float64) {
	durAry := strings.Split(dur, ":")
	var secs float64
	if len(durAry) != 3 {
		return
	}
	hr, _ := strconv.ParseFloat(durAry[0], 64)
	secs = hr * (60 * 60)
	min, _ := strconv.ParseFloat(durAry[1], 64)
	secs += min * (60)
	second, _ := strconv.ParseFloat(durAry[2], 64)
	secs += second
	return secs
}

// 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 创建文件夹
func CreateDirectory(dir string) (err error) {
	var exist bool
	exist, err = PathExists(dir)
	if err != nil {
		return
	}

	if !exist {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return
		}
	}
	return
}

// float64 保留2位小数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// 返回格式化 string
func StrVal(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

//字节数(大端)组转成int(无符号的)
func BytesToIntU(b []byte) (int64, error) {
	//字节转换成整形

	bitStr := string(b)

	sai, err := strconv.Atoi(bitStr)

	if err != nil {
		return 0, err
	}

	return int64(sai), nil

}
