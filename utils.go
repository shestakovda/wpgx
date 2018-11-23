package wpgx

// import (
// 	"database/sql"
// 	"database/sql/driver"
// 	"encoding/hex"
// 	"encoding/json"
// 	"encoding/xml"
// 	"fmt"
// 	"strings"
// 	"time"

// 	"github.com/satori/go.uuid"
// )

// // GUID - генерация нового GUID.V4
// func GUID() string {
// 	v4, _ := uuid.NewV4()
// 	return v4.String()
// }

// // UUID - генерация нового UUID.V4
// func UUID() string {
// 	v4, _ := uuid.NewV4()
// 	return hex.EncodeToString(v4.Bytes())
// }

// /**** Strings ****/

// // Strings - Реализация IList для простого списка строк
// type Strings []string

// // NewItem - создание нового элемента списка
// func (s *Strings) NewItem() IListItem { return new(String) }

// // Append - Добавление нового элемента
// func (s *Strings) Append(item IListItem) error {
// 	str, ok := item.(*String)
// 	if !ok {
// 		return fmt.Errorf("Неизвестный тип добавляемого элемента")
// 	}
// 	if str.Valid {
// 		*s = append(*s, str.Get())
// 	}
// 	return nil
// }

// // NewString - Конструктор для строки
// func NewString(data string) String {
// 	var s String
// 	s.SetValue(data)
// 	return s
// }

// // String - строка с возможностью NULL, элемент списка Strings
// type String struct{ sql.NullString }

// // Placeholder - реализация IListItem
// func (s *String) Placeholder(name string) interface{} { return &s.NullString }

// // Set - для установки значения
// func (s *String) Set(value string) {
// 	s.NullString.String = value
// 	s.NullString.Valid = value != ""
// }

// // Get - получение значения по-умолчанию
// func (s *String) Get() string {
// 	if s.Valid {
// 		return s.NullString.String
// 	}
// 	return ""
// }

// func (s String) String() string { return s.Get() }

// // IsNull - Проверка на наличие значения
// func (s String) IsNull() bool { return !s.NullString.Valid }

// // SetValue - Явная установка значения, в т.ч. пустого
// func (s *String) SetValue(str string) { s.NullString.Valid = true; s.NullString.String = str }

// // MarshalJSON - реализация json.Marshaler
// func (s String) MarshalJSON() ([]byte, error) {
// 	if !s.Valid {
// 		return []byte(`null`), nil
// 	}
// 	return json.Marshal(s.NullString.String)
// }

// // UnmarshalJSON - реализация json.Unmarshaler
// func (s *String) UnmarshalJSON(data []byte) (err error) {
// 	if data == nil {
// 		s.NullString.Valid = false
// 		return nil
// 	}
// 	s.NullString.Valid = true
// 	if err = json.Unmarshal(data, &s.NullString.String); err != nil {
// 		return
// 	}
// 	return nil
// }

// // MarshalXML - Реализация xml.Marshaler
// func (s String) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
// 	if !s.NullString.Valid {
// 		return e.EncodeElement(nil, start)
// 	}
// 	return e.EncodeElement(s.NullString.String, start)
// }

// // UnmarshalXML - Реализация xml.Unmarshaler
// func (s *String) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
// 	s.NullString.Valid = true
// 	if err = d.DecodeElement(&s.NullString.String, &start); err != nil {
// 		return
// 	}
// 	return nil
// }

// // MarshalXMLAttr - Реализация xml.MarshalerAttr
// func (s String) MarshalXMLAttr(name xml.Name) (a xml.Attr, err error) {
// 	if s.NullString.Valid {
// 		a.Name, a.Value = name, s.NullString.String
// 	}
// 	return
// }

// // UnmarshalXMLAttr - Реализация xml.UnmarshalerAttr
// func (s *String) UnmarshalXMLAttr(attr xml.Attr) error {
// 	s.SetValue(attr.Value)
// 	return nil
// }

// /***** Ints *******/

// // Ints - Реализация IList для простого списка целых чисел
// type Ints []int

// // NewItem - создание нового элемента списка
// func (i *Ints) NewItem() IListItem { return new(Int) }

// // Append - Добавление нового элемента
// func (i *Ints) Append(item IListItem) error {
// 	num, ok := item.(*Int)
// 	if !ok {
// 		return fmt.Errorf("Неизвестный тип добавляемого элемента")
// 	}
// 	if num.Valid {
// 		*i = append(*i, int(num.Int64))
// 	}
// 	return nil
// }

// // Int - целое (int64) с возможностью NULL, элемент списка Ints
// type Int struct{ sql.NullInt64 }

// // Placeholder - реализация IListItem
// func (i *Int) Placeholder(name string) interface{} { return &i.NullInt64 }

// // Get - получение значения по-умолчанию
// func (i *Int) Get() int64 {
// 	if i.Valid {
// 		return i.NullInt64.Int64
// 	}
// 	return 0
// }

// // MarshalJSON - реализация json.Marshaler
// func (i *Int) MarshalJSON() ([]byte, error) {
// 	if !i.Valid {
// 		return []byte(`null`), nil
// 	}
// 	return json.Marshal(i.NullInt64.Int64)
// }

// // UnmarshalJSON - реализация json.Unmarshaler
// func (i *Int) UnmarshalJSON(data []byte) (err error) {
// 	if data == nil {
// 		i.NullInt64.Valid = false
// 		return nil
// 	}
// 	i.NullInt64.Valid = true
// 	if err = json.Unmarshal(data, &i.NullInt64.Int64); err != nil {
// 		return
// 	}
// 	return nil
// }

// /******* Float *******/

// // NewFloat - Конструктор для дроби
// func NewFloat(value float64) Float { return new(Float).Set(value) }

// // Float - дробное (float64) с возможностью NULL
// type Float struct{ sql.NullFloat64 }

// // Placeholder - реализация IListItem
// func (f *Float) Placeholder(name string) interface{} { return &f.NullFloat64 }

// // Get - получение значения по-умолчанию
// func (f *Float) Get() float64 {
// 	if f.Valid {
// 		return f.NullFloat64.Float64
// 	}
// 	return 0
// }

// // Set - для установки значения
// func (f *Float) Set(value float64) Float {
// 	f.NullFloat64.Valid = true
// 	f.NullFloat64.Float64 = value
// 	return *f
// }

// // MarshalJSON - реализация json.Marshaler
// func (f *Float) MarshalJSON() ([]byte, error) {
// 	if !f.Valid {
// 		return []byte(`null`), nil
// 	}
// 	return json.Marshal(f.NullFloat64.Float64)
// }

// // UnmarshalJSON - реализация json.Unmarshaler
// func (f *Float) UnmarshalJSON(data []byte) (err error) {
// 	if data == nil {
// 		f.NullFloat64.Valid = false
// 		return nil
// 	}
// 	f.NullFloat64.Valid = true
// 	if err = json.Unmarshal(data, &f.NullFloat64.Float64); err != nil {
// 		return
// 	}
// 	return nil
// }

// /******** Times ********/

// // Times - Реализация IList для простого списка объектов даты/времени
// type Times []time.Time

// // NewItem - создание нового элемента списка
// func (t *Times) NewItem() IListItem { return new(Time) }

// // Append - Добавление нового элемента
// func (t *Times) Append(item IListItem) error {
// 	tm, ok := item.(*Time)
// 	if !ok {
// 		return fmt.Errorf("Неизвестный тип добавляемого элемента")
// 	}
// 	if tm.Valid {
// 		*t = append(*t, tm.Time)
// 	}
// 	return nil
// }

// // Time - объект даты/времени с возможностью NULL, элемент списка Times
// type Time struct {
// 	time.Time
// 	Valid bool
// }

// // Now - Заменитель time.Now() для нестандартного типа
// func Now() Time { return Time{Time: time.Now(), Valid: true} }

// // Placeholder - реализация IListItem
// func (t *Time) Placeholder(name string) interface{} { return t }

// // Scan - реализация Scanner
// func (t *Time) Scan(value interface{}) error {
// 	t.Time, t.Valid = value.(time.Time)
// 	if t.Valid {
// 		t.Valid = !t.Time.IsZero()
// 	}
// 	return nil
// }

// // Value - реализация Valuer
// func (t Time) Value() (driver.Value, error) {
// 	if !t.Valid {
// 		return nil, nil
// 	}
// 	return t.Time, nil
// }

// // IsNull - Проверка на наличие значения
// func (t Time) IsNull() bool { return !t.Valid }

// // Set - Явная установка значения
// func (t *Time) Set(n time.Time) { t.Valid = true; t.Time = n }

// // Get - Получение значения
// func (t *Time) Get() time.Time {
// 	if t.Valid {
// 		return t.Time
// 	}
// 	return time.Time{}
// }

// // MarshalJSON - реализация json.Marshaler
// func (t Time) MarshalJSON() ([]byte, error) {
// 	if !t.Valid {
// 		return []byte(`null`), nil
// 	}
// 	return json.Marshal(t.Time)
// }

// // UnmarshalJSON - реализация json.Unmarshaler
// func (t Time) UnmarshalJSON(data []byte) (err error) {
// 	if data == nil {
// 		t.Valid = false
// 		return nil
// 	}
// 	t.Valid = true
// 	if err = json.Unmarshal(data, &t.Time); err != nil {
// 		return
// 	}
// 	return nil
// }

// // RawList - обычный список для быстрых выборок
// type RawList []map[string]string

// // Шаблоны запросов на вставку
// const (
// 	TplInsertUpdate = `
// INSERT INTO "%s"."%s" ("%s")
// VALUES (%s)
// ON CONFLICT ("%s")
// DO UPDATE SET %s;
// `
// 	TplInsertUpdateReturning = `
// INSERT INTO "%s"."%s" ("%s")
// VALUES (%s)
// ON CONFLICT ("%s")
// DO UPDATE SET %s
// RETURNING "%s";
// `
// 	TplInsertIgnore = `
// INSERT INTO "%s"."%s" ("%s")
// VALUES (%s)
// ON CONFLICT ("%s")
// DO NOTHING;
// `
// 	TplUpdate = `
// UPDATE "%s"."%s" SET %s
// WHERE "%s" = %s;
// `
// )

// // InsertText - хелпер для построения текста запроса на вставку с обновлением
// func InsertText(schema string, table string, cols []string, keys []string, update bool, returnField string) (res string, names []string) {

// 	// Получение имен колонок и модификаторов типов
// 	mods := make(map[string]string, len(cols))
// 	names = make([]string, len(cols))
// 	for i := range cols {
// 		parts := strings.Split(cols[i], "::")
// 		names[i] = parts[0]
// 		if len(parts) > 1 {
// 			mods[names[i]] = parts[1]
// 		}
// 	}

// 	// Подготавливаем нумерованные замены
// 	places := make([]string, len(names))
// 	for i := 0; i < len(names); i++ {
// 		num := `$%d`
// 		if mod := mods[names[i]]; mod != "" {
// 			num += `::` + mod
// 		}
// 		places[i] = fmt.Sprintf(num, i+1)
// 	}

// 	// Получаем отдельные куски запроса
// 	sNames := strings.Join(names, `", "`)
// 	sPlaces := strings.Join(places, `, `)
// 	sKeys := strings.Join(keys, `", "`)

// 	if update {
// 		// Строки обновления при конфликте
// 		updates := make([]string, 0, len(names)-len(keys))
// 		for i := range names {
// 			found := false
// 			for j := range keys {
// 				if names[i] == keys[j] {
// 					found = true
// 					break
// 				}
// 			}
// 			if !found {
// 				updates = append(updates, fmt.Sprintf(`"%s" = excluded."%s"`, names[i], names[i]))
// 			}
// 		}
// 		sUpdates := strings.Join(updates, ",\n")

// 		if returnField == "" {
// 			res = fmt.Sprintf(TplInsertUpdate, schema, table, sNames, sPlaces, sKeys, sUpdates)
// 		} else {
// 			res = fmt.Sprintf(TplInsertUpdateReturning, schema, table, sNames, sPlaces, sKeys, sUpdates, returnField)
// 		}
// 	} else {
// 		res = fmt.Sprintf(TplInsertIgnore, schema, table, sNames, sPlaces, sKeys)
// 	}
// 	return
// }

// // UpdateText - хелпер для построения текста запроса на обновление
// func UpdateText(schema string, table string, cols []string, whereCol string) (res string, names []string) {
// 	// todo: добавить возможность множественных условия для обновления (может пригодиться, наверное)
// 	// Получение имен колонок и модификаторов типов
// 	// todo: проверить корректность в случае наличия модификаторов типов
// 	mods := make(map[string]string, len(cols))
// 	names = make([]string, len(cols))
// 	for i := range cols {
// 		parts := strings.Split(cols[i], "::")
// 		names[i] = parts[0]
// 		if len(parts) > 1 {
// 			mods[names[i]] = parts[1]
// 		}
// 	}

// 	// Подготавливаем нумерованные замены
// 	updates := make([]string, 0, len(names))
// 	for i := 0; i < len(names); i++ {
// 		num := `$%d`
// 		if mod := mods[names[i]]; mod != "" {
// 			num += `::` + mod
// 		}
// 		updates = append(updates, fmt.Sprintf(`"%s" = %s`, names[i], fmt.Sprintf(num, i+1)))
// 	}
// 	sUpdates := strings.Join(updates, ",\n")
// 	wherePlace := fmt.Sprintf(`$%d`, len(names)+1)
// 	// отдельно для where условия
// 	names = append(names, whereCol)

// 	res = fmt.Sprintf(TplUpdate, schema, table, sUpdates, whereCol, wherePlace)
// 	return
// }
