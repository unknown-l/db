package db

type fieldItem struct {
	alias string
	name  string
}

type field struct {
	string string       // 原始字符串
	item   []*fieldItem // 解析后的格式
}

func (f *field) Set(string string) {
	f.item = make([]*fieldItem, 0)
	brackets := 0
	item := &fieldItem{alias: "", name: ""}
	for i := 0; i < len(string); i++ {
		char := string[i : i+1]
		// 匹配到括号 重置name
		if char == "(" {
			brackets++
			item.name = ""
			continue
		}
		// 删除括号
		if char == ")" {
			brackets--
			continue
		}
		// 如果是.,找到alias,重置name
		if char == "." {
			item.alias = string[i-1 : i]
			item.name = ""
			continue
		}
		// 如果是空格重置name
		if char == " " {
			item.name = ""
			continue
		}
		// 找到没在括号中的逗号，就是下一个field
		if brackets == 0 && char == "," {
			// 没有写alias,默认主表字段
			if item.alias == "" {
				item.alias = tableMasterAlias
			}
			f.item = append(f.item, item)
			item = &fieldItem{alias: "", name: ""}
			continue
		}
		item.name += char
	}
	// 没有写alias,默认主表字段
	if item.alias == "" {
		item.alias = tableMasterAlias
	}
	f.item = append(f.item, item)
	f.string = string
}
