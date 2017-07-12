package logo

import (
	"fmt"
	"os"
	"cloud.ssh/config"
)

func Show()  {
	port := config.Read("socket", "port")
	str := `│＼＿＿╭╭╭╭╭＿＿／│
│　　　　　　　　 │
│　　　　　　　　 │
│　＞　　　　　●　│
│≡　　╰┬┬┬╯　　≡  │
│　　　　╰—╯　　  │
╰——┬ｏ——————ｏ┬——╯
　　│ Cloud SSH│
　　╰┬————————┬╯ `
	fmt.Println(str)
	fmt.Printf("\n%c[1;40;32m%s%c[0m\n\n", 0x1B, "Welcome, Cloud SSH Start Success", 0x1B)
	fmt.Printf("%c[1;40;32m%s%d%c[0m\n", 0x1B, "Process ID : ", os.Getpid() ,0x1B)
	fmt.Printf("%c[1;40;32m%s%s%c[0m\n", 0x1B, "Port       : ", port, 0x1B)
	fmt.Printf("%c[1;40;32m%s%c[0m\n", 0x1B, "Author     : 夜云", 0x1B)
	fmt.Printf("%c[1;40;32m%s%c[0m\n\n", 0x1B, "GitHub     : https://github.com/aiyeyun", 0x1B)
}
