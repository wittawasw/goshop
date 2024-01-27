# สร้าง path แบบ dynamic ด้วย Gorilla/Mux

สมมติว่าเราต้องการสร้างสร้างเว็บที่มี URL
- `/` - หน้าหลักที่มีสินค้าทั้งหมด
- `/products/:id` - คือหน้าแยกของสินค้าแต่ล่ะชิ้น

ใน template ขั้นต้น ยังไม่สามารถทำ `/products/:id` ได้ เราจึงต้องอาศัย library `gorilla/mux` มาทำในส่วนนี้

## 1: ใช้คำสั่ง `go get` เพื่อ install Gorilla/Mux
```bash
# -u ย่อมาจาก update หมายถึงถ้ามีอยู่แล้วแต่ไม่ใช่เวอร์ชันล่าสุด ก็ให้ดาวน์โหลดมาอัพเดต
go get -u github.com/gorilla/mux
```

## 2: แยกไฟล์ routes ออกมาจาก server.go

### สร้างไฟล์สำหรับเก็บ routes
- สร้างไฟล์ที่ `web/routes.go` โดยไฟล์นี้จะเป็นไฟล์ที่ใช้รวม routes ต่างๆ แทน `server.go`
- เพิ่ม `productHandler`
  ```go
  package web

  import (
      "net/http"
      "github.com/gorilla/mux"
  )

  func productHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    w.Write([]byte("Product Details for ID: " + id))
  }
  ```
### แก้ `assetHandler`
- เนื่องจากเราจะทำการตั้งค่า routes ด้วย `gorilla/mux` จึงทำให้ `assetHandler` จะไม่สามารถใช้ต่อได้ จึงต้องทำการแก้ไข code ภายในให้เป็นไปตามรูปแบบของ `gorilla/mux`
  ```go
  func assetsHandler(dir string, prefix string) http.Handler {
    return http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))
  }
  ```
### ย้าย handler ทั้งหมดมาไว้ที่ `web/routes.go`
- สร้าง `func` เพื่อเป็นที่รวมและให้ `web/server.go` สารถเรียกใช้ได้สะดวก
  ```go
  func RegisterRoutes(r *mux.Router) {
    r.HandleFunc("/", indexHandler)
    // แก้การ register ตาม assetsHandler ที่เขียนขึ้นใหม่
    r.PathPrefix("/assets/").Handler(assetsHandler("web/assets/", "/assets/"))
    r.HandleFunc("/products/{id:[0-9]+}", productHandler)
  }
  ```
## 3: ผลลัพธ์สุดท้าย
  ```go
  // web/routes.go
  package web

  import (
    "html/template"
    "net/http"

    "github.com/gorilla/mux"
  )

  func RegisterRoutes(r *mux.Router) {
    r.HandleFunc("/", indexHandler)
    r.PathPrefix("/assets/").Handler(assetsHandler("web/assets/", "/assets/"))
    r.HandleFunc("/products/{id:[0-9]+}", productHandler)
  }

  // Serve Homepage
  // path: /
  func indexHandler(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("web/templates/index.html")
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    data := struct {
      Title string
    }{
      Title: "Homepage",
    }

    if err := tmpl.Execute(w, data); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }

  // Serve Static Files
  // path: /assets
  func assetsHandler(dir string, prefix string) http.Handler {
    return http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))
  }

  // Serve Product page
  // path: /product/:id
  func productHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    w.Write([]byte("Product Details for ID: " + id))
  }

  ```

  ```go
  // web/server.go
  package web

  import (
    "fmt"
    "net/http"

    "github.com/gorilla/mux"
  )

  // Start an HTTP Server with Handlers
  func SetupAndServe() {
    // ส่วนที่ gorilla/mux มาแทนที่
    r := mux.NewRouter()

    // ใช้ package web เหมือนกันจึงสามารถใช้ RegisterRoutes ได้โดยตรง
    // ถ้าให้ web/routes.go เป็น package routes ก็จะใช้ routes.RegisterRoutes(r)
    RegisterRoutes(r)

    fmt.Println("Server started on http://localhost:8080")
    http.ListenAndServe(":8080", r)
  }
  ```

## 4: ลองทำเองต่อ
- ทำให้ `productHandler` เป็นการ render HTML template โดยส่งข้อมูล เช่น name, price
ไปที่ template
- ดูตัวอย่างการ render HTML template จาก `indexHandler`



# สร้าง struct ของ Product ไว้ใช้งาน
เราต้องการ struct ของ product เพื่อให้สะดวกในการอ้างอิงเวลาส่งต่อให้ func ต่างๆ

### 1: สร้างไฟล์ `/internal/models/product.go`
```go
package models

type Product struct {
	ID    string
	Name  string
	Price float64
}

// Mock sample products
var ProductList = []Product{
	{ID: "1", Name: "Product A", Price: 19.99},
	{ID: "2", Name: "Product B", Price: 29.99},
	{ID: "3", Name: "Product C", Price: 39.99},
}
```

### 2: นำไปใช้ใน `productHandler`
```go
func productHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// หา product จาก product ID
	var product models.Product
	for _, p := range models.ProductList {
		if p.ID == id {
			product = p
			break
		}
	}

	if product.ID == "" {
		http.NotFound(w, r)
		return
	}

	// ใช้การ render HTML template จากก่อนหน้านี้ร่วมด้วย
	response := fmt.Sprintf("Product Details for ID: %s\nName: %s\nPrice: %.2f", product.ID, product.Name, product.Price)
	w.Write([]byte(response))
}
```

## 3: ลองทำเองต่อ
- ลองกำหนดค่าต่างๆเพิ่มเติม ให้ `Product` และ render ด้วย HTML template
- ลองหาวิธีสร้าง link ของ product ทั้งหมดภายใน `ProductList` โดยให้ render ไปที่ homepage
