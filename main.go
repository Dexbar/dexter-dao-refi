package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Ahora la base de datos es el perfil de UN usuario
type UserProfile struct {
	Balance         int
	Inventario      []string
	Historial       []string
	VotosRobot      int
	VotosPago       int
	DiscordWebhook  string
	ContractAddress string
	TokenAddress    string
	UltimoReclamo   int64
}

// MarketItem represents an NFT in the marketplace
type MarketItem struct {
	Key         string // internal identifier used in code (e.g., "ajolote")
	ImgPath     string // URL path to serve the image
	DisplayName string // Human readable name shown in the UI
	Price       int    // Price in DXT tokens
	Name        string // Same as Key, kept for convenience
	MaxSupply   int    // Maximum allowed copies
	Rarity      string // Rarity tier (Común, Raro, Épico, Legendario, Único)
	Description string // NFT description
	Vendidos    int    // Number of sold copies
	Elemento    string // NFT element type
	Poder       int    // NFT power level
	RarezaPorc  int    // NFT rarity percentage
}

var marketItems = []MarketItem{{
	Key:         "ajolote",
	ImgPath:     "/img-maestro",
	DisplayName: "Maestro #001",
	Price:       200,
	Name:        "ajolote",
	MaxSupply:   500,
	Rarity:      "Común",
	Description: "NFT de colección — Ajolote Maestro del DAO",
	Elemento:    "Agua",
	Poder:       75,
	RarezaPorc:  80,
}, {
	Key:         "luna",
	ImgPath:     "/img-astronauta",
	DisplayName: "Astronauta",
	Price:       500,
	Name:        "luna",
	MaxSupply:   100,
	Rarity:      "Raro",
	Description: "Explorador del cosmos — Edición Astronauta",
	Elemento:    "Cósmico",
	Poder:       85,
	RarezaPorc:  90,
}, {
	Key:         "quetzal",
	ImgPath:     "/img-quetzal",
	DisplayName: "Dios Quetzal",
	Price:       1000,
	Name:        "quetzal",
	MaxSupply:   25,
	Rarity:      "Épico",
	Description: "Deidad suprema — Quetzalcóatl Digital",
	Elemento:    "Divino",
	Poder:       95,
	RarezaPorc:  96,
}, {
	Key:         "androide",
	ImgPath:     "/img-androide",
	DisplayName: "Cyber Androide",
	Price:       2000,
	Name:        "androide",
	MaxSupply:   5,
	Rarity:      "Legendario",
	Description: "Cyber Androide — Edición Legendaria Exclusiva",
	Elemento:    "Cyber",
	Poder:       98,
	RarezaPorc:  99,
}, {
	Key:         "supremo",
	ImgPath:     "/img-supremo",
	DisplayName: "Ajolote Supremo",
	Price:       5000,
	Name:        "supremo",
	MaxSupply:   5,
	Rarity:      "Limitado (5)",
	Description: "Soberano absoluto de la blockchain. Una sola copia existente en el universo.",
	Elemento:    "Cósmico/Oro",
	Poder:       100,
	RarezaPorc:  100,
}, {
	Key:         "chinampero",
	ImgPath:     "/img-chinampero",
	DisplayName: "Ajolote Chinampero",
	Price:       1500,
	Name:        "chinampero",
	MaxSupply:   200,
	Rarity:      "Épico",
	Description: "Protector de las chinampas tradicionales y guardián de la biodiversidad de Xochimilco.",
	Elemento:    "Tierra/Planta",
	Poder:       92,
	RarezaPorc:  95,
}}

func manejadorMarket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

    // Load the current user profile (same logic as manejadorPrincipal)
    cookie, err := r.Cookie("sesion")
    if err != nil || cookie.Value == "" {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    wallet := cookie.Value
    db := cargarDB()
    perfil, existe := db[wallet]
    if !existe {
        perfil = UserProfile{Balance: 0, Inventario: []string{}, Historial: []string{}}
        agregarHistorial(&perfil, "Cuenta Web3 creada exitosamente.")
    }
	mensajeNotif := ""
	if r.Method == "POST" && r.FormValue("buy") != "" {
		key := r.FormValue("buy")
		// Find the item
		for _, it := range marketItems {
			if it.Key == key {
				vendidos := obtenerCantidadVendida(db, key)
				if vendidos >= it.MaxSupply {
					mensajeNotif = "¡Sold out! Todas las copias de esta tarjeta han sido adquiridas."
				} else if perfil.Balance >= it.Price {
					perfil.Balance -= it.Price
					perfil.Inventario = append(perfil.Inventario, it.Name)
					agregarHistorial(&perfil, fmt.Sprintf("-%d TK Compra de %s", it.Price, it.Name))
					mensajeNotif = "¡NFT transferido a tu Bóveda!"
					
					var color int = 0x00ff00
					if it.Key == "supremo" {
						color = 16766720 // Oro
					}
					enviarDiscord(perfil.DiscordWebhook, "🎉 ¡NFT ADQUIRIDO!", fmt.Sprintf("Billetera: `%s` ha adquirido **%s** por **%d TK**.\n\n¡El NFT ha sido transferido a su bóveda!", wallet[:6]+"..."+wallet[len(wallet)-4:], it.DisplayName, it.Price), color, it.ImgPath)
				} else {
					mensajeNotif = "¡Saldo insuficiente!"
				}
				break
			}
		}
		// Save profile changes
		db[wallet] = perfil
		guardarDB(db)
	}

	// Dynamic calculation of sold amounts before rendering
	var renderedItems []MarketItem
	for _, it := range marketItems {
		it.Vendidos = obtenerCantidadVendida(db, it.Key)
		renderedItems = append(renderedItems, it)
	}

	// Próximo drop mensual: 1er día del mes siguiente a las 00:00:00
	ahoraTime := time.Now()
	año := ahoraTime.Year()
	mes := ahoraTime.Month()
	siguienteMes := mes + 1
	siguienteAño := año
	if siguienteMes > 12 {
		siguienteMes = 1
		siguienteAño++
	}
	proximoDrop := time.Date(siguienteAño, siguienteMes, 1, 0, 0, 0, 0, ahoraTime.Location())
	proximoDropUnix := proximoDrop.Unix()

	addrs := cargarDirecciones()

	// Render the marketplace page
	data := struct {
		Items           []MarketItem
		Year            int
		Mensaje         string
		ProximoDropUnix int64
		TokenAddress    string
		NFTAddress      string
		Network         string
	}{
		Items:           renderedItems,
		Year:            time.Now().Year(),
		Mensaje:         mensajeNotif,
		ProximoDropUnix: proximoDropUnix,
		TokenAddress:    addrs.DexterDAO,
		NFTAddress:      addrs.DexterNFT,
		Network:         addrs.Network,
	}
    tmpl, _ := template.ParseFiles("market.html")
    tmpl.Execute(w, data)
}


    // Duplicate struct definitions removed

// Y la base de datos global es un Mapa (Diccionario) de Wallet -> Perfil
type Database map[string]UserProfile

type PageData struct {
	User            string
	WalletCompleta  string
	Balance         int
	ValorPesos      int
	Mensaje         string
	Historial       []string
	VotosRobot      int
	VotosPago       int
	TieneAjolote    bool
	TieneLuna       bool
	TieneQuetzal    bool
	TieneAndroide   bool
	TieneSupremo    bool
	TieneChinampero bool
	DiscordWebhook  string
	TokenAddress    string
	NFTAddress      string
	GovAddress      string
	ContractAddress string
	UltimoReclamo   int64
	TiempoRestante  int64
	Network         string
}

type DeployedAddresses struct {
	DexterDAO string `json:"DexterDAO"`
	DexterNFT string `json:"DexterNFT"`
	DexterGov string `json:"DexterGov"`
	Network   string `json:"network"`
}

func cargarDirecciones() DeployedAddresses {
	datos, err := os.ReadFile("deployed_addresses.json")
	var addr DeployedAddresses
	if err == nil {
		json.Unmarshal(datos, &addr)
	}
	return addr
}

func mintearTokensBlockchain(network string, recipient string, amount int) (string, error) {
	netParam := network
	if netParam == "" {
		netParam = "hardhat"
	}
	if netParam == "hardhat" {
		netParam = "localhost"
	}

	cmd := exec.Command("node", "scripts/mint_tokens.js", recipient, strconv.Itoa(amount))
	cmd.Env = append(os.Environ(),
		"HARDHAT_NETWORK="+netParam,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	// Buscar confirmacion en la salida antes de validar errores de ejecucion (ej. crashes de libuv al salir)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "MINT_SUCCESS:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	if err != nil {
		return "", fmt.Errorf("ejecucion fallida: %v, stderr: %s, stdout: %s", err, stderr.String(), output)
	}

	return "", fmt.Errorf("no se encontro confirmacion en salida: %s", output)
}


type ConfirmNFTRequest struct {
	Wallet string `json:"wallet"`
	Key    string `json:"key"`
	TxHash string `json:"txHash"`
}

type ConfirmVoteRequest struct {
	Wallet     string `json:"wallet"`
	ProposalID int    `json:"proposalId"`
	TxHash     string `json:"txHash"`
}


const RUTA_MAESTRO = `assets/ajolote_maestro.png`
const RUTA_ASTRONAUTA = `assets/ajolote_astronauta.png`
const RUTA_QUETZAL = `assets/ajolote_quetzal.png`
const RUTA_ANDROIDE = `assets/ajolote_androide.png`
const RUTA_SUPREMO = `assets/ajolote_supremo.png`
const RUTA_CHINAMPERO = `assets/ajolote_chinampero.png`

func enviarDiscord(webhookURL string, titulo string, descripcion string, color int, imagePath string) {
	if webhookURL == "" {
		return
	}
	go func() {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)

		var embedImage map[string]interface{}
		var fileName string
		if imagePath != "" {
			fileName = filepath.Base(imagePath)
			embedImage = map[string]interface{}{
				"url": "attachment://" + fileName,
			}
		}

		embed := map[string]interface{}{
			"title":       titulo,
			"description": descripcion,
			"color":       color,
			"footer": map[string]interface{}{
				"text": "Dexter DAO - Powered by Go & Web3",
			},
		}
		if imagePath != "" {
			embed["image"] = embedImage
		}

		payload := map[string]interface{}{
			"username":   "DEXTER DAO Bot",
			"avatar_url": "https://img.icons8.com/color/96/ethereum.png",
			"embeds":     []interface{}{embed},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return
		}

		_ = w.WriteField("payload_json", string(payloadBytes))

		if imagePath != "" {
			file, err := os.Open(imagePath)
			if err == nil {
				defer file.Close()
				part, err := w.CreateFormFile("files[0]", fileName)
				if err == nil {
					_, _ = io.Copy(part, file)
				}
			}
		}

		w.Close()

		req, err := http.NewRequest("POST", webhookURL, &b)
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}()
}


func cargarDB() Database {
	datos, err := os.ReadFile("db.json")
	var db Database
	if err == nil {
		err = json.Unmarshal(datos, &db)
		// Si el archivo viejo no es compatible con el nuevo formato de mapa, creamos uno nuevo
		if err != nil { db = make(Database) }
	} else {
		db = make(Database)
	}
	return db
}

func guardarDB(db Database) {
	datosJSON, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		fmt.Printf("Error al serializar DB: %v\n", err)
		return
	}
	err = os.WriteFile("db.json", datosJSON, 0644)
	if err != nil {
		fmt.Printf("Error al escribir db.json en disco: %v\n", err)
	} else {
		fmt.Println("db.json guardada exitosamente en disco")
	}
}

func agregarHistorial(perfil *UserProfile, msg string) {
	fecha := time.Now().Format("02/01 15:04:05")
	registro := fmt.Sprintf("[%s] %s", fecha, msg)
	perfil.Historial = append([]string{registro}, perfil.Historial...)
}

func tieneItem(inv []string, item string) bool {
	for _, i := range inv {
		if i == item { return true }
	}
	return false
}

func quitarItem(inv []string, item string) []string {
	var nuevo []string
	for _, i := range inv {
		if i != item { nuevo = append(nuevo, i) }
	}
	return nuevo
}

func obtenerCantidadVendida(db Database, key string) int {
	count := 0
	for _, perfil := range db {
		for _, item := range perfil.Inventario {
			if item == key {
				count++
			}
		}
	}
	return count
}

func manejadorPrincipal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	cookie, err := r.Cookie("sesion")
	
	if r.FormValue("accion") == "logout" {
		http.SetCookie(w, &http.Cookie{Name: "sesion", Value: "", MaxAge: -1})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Si no hay cookie de billetera, mandarlo al login
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	wallet := cookie.Value
	displayWallet := wallet
	if len(wallet) > 10 {
		displayWallet = wallet[:6] + "..." + wallet[len(wallet)-4:]
	}
	db := cargarDB()
	addrs := cargarDirecciones()
	
	// Buscar el perfil de esta billetera. Si no existe, crear uno virgen.
	perfil, existe := db[wallet]
	if !existe {
		perfil = UserProfile{Balance: 0, Inventario: []string{}, Historial: []string{}}
		agregarHistorial(&perfil, "Cuenta Web3 creada exitosamente.")
	}

	mensajeNotif := "¡Conexión Web3 establecida!"

	if r.Method == "POST" {
		if r.FormValue("recompensa_diaria") != "" {
			ahora := time.Now().Unix()
			elapsed := ahora - perfil.UltimoReclamo
			if perfil.UltimoReclamo == 0 || elapsed >= 86400 {
				perfil.Balance += 100
				perfil.UltimoReclamo = ahora
				agregarHistorial(&perfil, "+100 TK Recompensa Diaria Reclamada")
				mensajeNotif = "¡Recompensa diaria reclamada con éxito! +100 TK"
				enviarDiscord(perfil.DiscordWebhook, "🎁 ¡RECOMPENSA DIARIA RECLAMADA!", fmt.Sprintf("Billetera: `%s` ha reclamado su recompensa diaria de **100 TK**.\n\n¡Saldo actualizado: **%d TK**!", displayWallet, perfil.Balance), 65280, "") // Verde
			} else {
				restante := 86400 - elapsed
				horas := restante / 3600
				minutos := (restante % 3600) / 60
				mensajeNotif = fmt.Sprintf("¡Aún no puedes reclamar! Restan %dh %dm", horas, minutos)
			}
		}
		if r.FormValue("actualizar_webhook") != "" {
			perfil.DiscordWebhook = r.FormValue("webhook_url")
			if perfil.DiscordWebhook != "" {
				agregarHistorial(&perfil, "Webhook de Discord configurado.")
				mensajeNotif = "¡Webhook de Discord guardado!"
				enviarDiscord(perfil.DiscordWebhook, "🔌 ¡DEXTER DAO CONECTADO!", fmt.Sprintf("¡Servidor de Go conectado exitosamente a este canal!\nBilletera: `%s`", displayWallet), 65280, "") // Verde
			} else {
				agregarHistorial(&perfil, "Webhook de Discord removido.")
				mensajeNotif = "¡Webhook de Discord eliminado!"
			}
		}

		if r.FormValue("actualizar_contrato") != "" {
			perfil.ContractAddress = r.FormValue("contract_address")
			if perfil.ContractAddress != "" {
				agregarHistorial(&perfil, "Contrato inteligente vinculado: "+perfil.ContractAddress)
				mensajeNotif = "¡Contrato inteligente actualizado!"
			} else {
				agregarHistorial(&perfil, "Contrato inteligente removido.")
				mensajeNotif = "¡Contrato inteligente eliminado!"
			}
		}

		if r.FormValue("horas") != "" {
			horas, _ := strconv.Atoi(r.FormValue("horas"))
			if horas > 0 {
				basePago := 10
				bonoMsg := ""
				if tieneItem(perfil.Inventario, "chinampero") {
					basePago = 15
					bonoMsg = " (Bono 1.5x Chinampero Activo! 🌾)"
				}
				pago := horas * basePago

				// Intentar mintear on-chain
				if addrs.DexterDAO == "" {
					mensajeNotif = "Error: El token DXT no ha sido desplegado todavía."
				} else {
					txHash, err := mintearTokensBlockchain(addrs.Network, wallet, pago)
					if err != nil {
						mensajeNotif = fmt.Sprintf("Error al minar on-chain: %v", err)
					} else {
						perfil.Balance += pago
						agregarHistorial(&perfil, fmt.Sprintf("+%d DXT minados on-chain (Tx: %s)%s", pago, txHash, bonoMsg))
						mensajeNotif = fmt.Sprintf("¡Chamba registrada on-chain! +%d DXT%s (Tx: %s)", pago, bonoMsg, txHash[:10]+"...")
						
						txLink := fmt.Sprintf("https://sepolia.etherscan.io/tx/%s", txHash)
						if addrs.Network == "alfajores" {
							txLink = fmt.Sprintf("https://celo-alfajores.blockscout.com/tx/%s", txHash)
						} else if addrs.Network == "hardhat" || addrs.Network == "localhost" {
							txLink = "Red Local (Hardhat)"
						}
						
						enviarDiscord(perfil.DiscordWebhook, "⛏️ Chamba Minera Realizada On-Chain", fmt.Sprintf("Billetera: `%s`\n**Horas trabajadas:** %d hrs\n**Tokens minados:** +%d DXT%s\n**Transacción:** %s\n**Saldo actual:** %d DXT", displayWallet, horas, pago, bonoMsg, txLink, perfil.Balance), 16766720, "") // Oro
					}
				}
			}
		}

		if r.FormValue("votar") != "" {
			opcion := r.FormValue("votar")
			var propuesta string
			if opcion == "robot" {
				perfil.VotosRobot++
				propuesta = "🤖 NUEVO NFT: ROBOT"
			}
			if opcion == "pago" {
				perfil.VotosPago++
				propuesta = "💰 SUBIR PAGO"
			}
			mensajeNotif = "¡Tu voto ha sido firmado en la blockchain simulada!"
			enviarDiscord(perfil.DiscordWebhook, "🗳️ VOTO DE GOBERNANZA FIRMADO", fmt.Sprintf("Billetera: `%s` ha emitido un voto para la propuesta:\n**%s**", displayWallet, propuesta), 16753920, "") // Naranja
		}

		if r.FormValue("comprar") != "" {
			it := r.FormValue("comprar")
			p := 0
			var imgPath string
			var nftNombre string
			var color int
			switch it {
			case "ajolote":
				p = 200
				imgPath = RUTA_MAESTRO
				nftNombre = "👾 Maestro #001"
				color = 16711935 // Magenta
			case "luna":
				p = 500
				imgPath = RUTA_ASTRONAUTA
				nftNombre = "⭐ Astronauta"
				color = 65535 // Cyan
			case "quetzal":
				p = 1000
				imgPath = RUTA_QUETZAL
				nftNombre = "👑 Dios Quetzal"
				color = 16766720 // Oro
			case "androide":
				p = 2000
				imgPath = RUTA_ANDROIDE
				nftNombre = "🤖 Cyber Androide"
				color = 65280 // Verde Neón
			case "supremo":
				p = 5000
				imgPath = RUTA_SUPREMO
				nftNombre = "👑 Ajolote Supremo"
				color = 16766720 // Oro
			case "chinampero":
				p = 1500
				imgPath = RUTA_CHINAMPERO
				nftNombre = "🌾 Ajolote Chinampero"
				color = 65280 // Verde
			}
			
			// Find the item definition for supply limit check
			var itemDef MarketItem
			found := false
			for _, item := range marketItems {
				if item.Key == it {
					itemDef = item
					found = true
					break
				}
			}
			
			if found {
				vendidos := obtenerCantidadVendida(db, it)
				if vendidos >= itemDef.MaxSupply {
					mensajeNotif = "¡Sold out! Todas las copias de esta tarjeta han sido adquiridas."
				} else if perfil.Balance >= p {
					perfil.Balance -= p
					perfil.Inventario = append(perfil.Inventario, it)
					agregarHistorial(&perfil, fmt.Sprintf("-%d TK Compra de %s", p, it))
					mensajeNotif = "¡NFT transferido a tu Bóveda!"
					enviarDiscord(perfil.DiscordWebhook, "🎉 ¡NFT ADQUIRIDO!", fmt.Sprintf("Billetera: `%s` ha adquirido **%s** por **%d TK**.\n\n¡El NFT ha sido transferido a su bóveda!", displayWallet, nftNombre, p), color, imgPath)
				} else {
					mensajeNotif = "¡Saldo insuficiente!"
				}
			}
		}

		if r.FormValue("vender") != "" {
			it := r.FormValue("vender")
			v := 0
			var imgPath string
			var nftNombre string
			var color int
			switch it {
			case "ajolote":
				v = 250
				imgPath = RUTA_MAESTRO
				nftNombre = "👾 Maestro #001"
				color = 16711680 // Rojo
			case "luna":
				v = 600
				imgPath = RUTA_ASTRONAUTA
				nftNombre = "⭐ Astronauta"
				color = 16711680 // Rojo
			case "quetzal":
				v = 1500
				imgPath = RUTA_QUETZAL
				nftNombre = "👑 Dios Quetzal"
				color = 16711680 // Rojo
			case "androide":
				v = 3000
				imgPath = RUTA_ANDROIDE
				nftNombre = "🤖 Cyber Androide"
				color = 16711680 // Rojo
			case "supremo":
				v = 7500
				imgPath = RUTA_SUPREMO
				nftNombre = "👑 Ajolote Supremo"
				color = 16711680 // Rojo
			case "chinampero":
				v = 2000
				imgPath = RUTA_CHINAMPERO
				nftNombre = "🌾 Ajolote Chinampero"
				color = 16711680 // Rojo
			}
			
			if tieneItem(perfil.Inventario, it) {
				perfil.Balance += v
				perfil.Inventario = quitarItem(perfil.Inventario, it)
				agregarHistorial(&perfil, fmt.Sprintf("+%d TK Venta de %s", v, it))
				mensajeNotif = fmt.Sprintf("¡Flipping exitoso! Ganaste %d TK", v)
				enviarDiscord(perfil.DiscordWebhook, "💸 NFT VENDIDO (FLIPPING)", fmt.Sprintf("Billetera: `%s` ha vendido su **%s** en el mercado secundario por **%d TK**.\n\n¡Ganancia realizada exitosamente!", displayWallet, nftNombre, v), color, imgPath)
			}
		}
		
		// Guardar los cambios del perfil en la base de datos general
		db[wallet] = perfil
		guardarDB(db)
	}

	// La dirección ya está truncada en displayWallet al inicio de la función

	ahora := time.Now().Unix()
	tiempoRestante := int64(0)
	if perfil.UltimoReclamo > 0 {
		elapsed := ahora - perfil.UltimoReclamo
		if elapsed < 86400 {
			tiempoRestante = 86400 - elapsed
		}
	}

	// Si el contrato guardado es diferente al token DXT desplegado, lo actualizamos al correcto para evitar errores
	if addrs.DexterDAO != "" && perfil.ContractAddress != addrs.DexterDAO {
		perfil.ContractAddress = addrs.DexterDAO
		db[wallet] = perfil
	}

	datosHTML := PageData{
		User:            displayWallet,
		WalletCompleta:  wallet,
		Balance:         perfil.Balance,
		ValorPesos:      perfil.Balance * 2,
		Mensaje:         mensajeNotif,
		Historial:       perfil.Historial,
		VotosRobot:      perfil.VotosRobot,
		VotosPago:       perfil.VotosPago,
		TieneAjolote:    tieneItem(perfil.Inventario, "ajolote"),
		TieneLuna:       tieneItem(perfil.Inventario, "luna"),
		TieneQuetzal:    tieneItem(perfil.Inventario, "quetzal"),
		TieneAndroide:   tieneItem(perfil.Inventario, "androide"),
		TieneSupremo:    tieneItem(perfil.Inventario, "supremo"),
		TieneChinampero: tieneItem(perfil.Inventario, "chinampero"),
		DiscordWebhook:  perfil.DiscordWebhook,
		TokenAddress:    addrs.DexterDAO,
		NFTAddress:      addrs.DexterNFT,
		GovAddress:      addrs.DexterGov,
		ContractAddress: perfil.ContractAddress,
		UltimoReclamo:   perfil.UltimoReclamo,
		TiempoRestante:  tiempoRestante,
		Network:         addrs.Network,
	}

	tmpl, _ := template.ParseFiles("index.html")
	tmpl.Execute(w, datosHTML)
}

func manejadorConfirmNFT(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req ConfirmNFTRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	wallet := strings.ToLower(req.Wallet)
	db := cargarDB()
	perfil, existe := db[wallet]
	if !existe {
		perfil = UserProfile{Balance: 0, Inventario: []string{}, Historial: []string{}}
	}

	// Encontrar el item en la lista para obtener el precio y nombre real
	var itemDef MarketItem
	encontrado := false
	for _, item := range marketItems {
		if item.Key == req.Key {
			itemDef = item
			encontrado = true
			break
		}
	}

	if !encontrado {
		http.Error(w, "NFT no encontrado", http.StatusNotFound)
		return
	}

	// Restar el balance local (simulado) para que coincida, o solo registrar si deseamos.
	if perfil.Balance >= itemDef.Price {
		perfil.Balance -= itemDef.Price
	} else {
		perfil.Balance = 0 // Si fue comprado directamente con fondos on-chain, sincronizamos
	}

	perfil.Inventario = append(perfil.Inventario, itemDef.Name)
	msgHistorial := fmt.Sprintf("-%d TK Compra en Blockchain de %s (Tx: %s)", itemDef.Price, itemDef.Name, req.TxHash)
	agregarHistorial(&perfil, msgHistorial)
	
	// Guardar en la DB
	db[wallet] = perfil
	guardarDB(db)

	// Mandar alerta a Discord
	var color int = 0x00ff00
	if req.Key == "supremo" {
		color = 16766720 // Oro
	} else if req.Key == "chinampero" {
		color = 0x00ff66 // Verde Chinampa
	}
	
	displayWallet := wallet
	if len(wallet) > 10 {
		displayWallet = wallet[:6] + "..." + wallet[len(wallet)-4:]
	}

	txLink := fmt.Sprintf("https://sepolia.etherscan.io/tx/%s", req.TxHash)
	if req.TxHash == "local" || len(req.TxHash) < 10 {
		txLink = "Red Local (Hardhat)"
	}

	enviarDiscord(
		perfil.DiscordWebhook,
		"⛓️ ¡NFT ACUÑADO EN BLOCKCHAIN REAL!",
		fmt.Sprintf("Billetera: `%s` ha acuñado con éxito **%s** por **%d DXT**.\n\n**Transacción:** %s\n\n¡El NFT ahora reside en la blockchain y en su bóveda!", displayWallet, itemDef.DisplayName, itemDef.Price, txLink),
		color,
		itemDef.ImgPath,
	)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("confirmado"))
}

func manejadorConfirmVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req ConfirmVoteRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	wallet := strings.ToLower(req.Wallet)
	db := cargarDB()
	perfil, existe := db[wallet]
	if !existe {
		perfil = UserProfile{Balance: 0, Inventario: []string{}, Historial: []string{}}
	}

	var propuesta string
	if req.ProposalID == 0 {
		perfil.VotosRobot++
		propuesta = "🤖 NUEVO NFT: ROBOT"
	} else if req.ProposalID == 1 {
		perfil.VotosPago++
		propuesta = "💰 SUBIR PAGO"
	} else {
		propuesta = fmt.Sprintf("Propuesta #%d", req.ProposalID)
	}

	msgHistorial := fmt.Sprintf("Voto firmado en Blockchain: %s (Tx: %s)", propuesta, req.TxHash)
	agregarHistorial(&perfil, msgHistorial)

	// Guardar en la DB
	db[wallet] = perfil
	guardarDB(db)

	displayWallet := wallet
	if len(wallet) > 10 {
		displayWallet = wallet[:6] + "..." + wallet[len(wallet)-4:]
	}

	txLink := fmt.Sprintf("https://sepolia.etherscan.io/tx/%s", req.TxHash)
	if req.TxHash == "local" || len(req.TxHash) < 10 {
		txLink = "Red Local (Hardhat)"
	}

	enviarDiscord(
		perfil.DiscordWebhook,
		"🗳️ VOTO DE GOBERNANZA REAL EN BLOCKCHAIN",
		fmt.Sprintf("Billetera: `%s` ha emitido y firmado su voto on-chain para:\n**%s**.\n\n**Transacción:** %s", displayWallet, propuesta, txLink),
		16753920, // Naranja
		"",
	)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("confirmado"))
}

func manejadorLogin(w http.ResponseWriter, r *http.Request) {
	// Recibir la billetera desde el frontend (JavaScript)
	if r.Method == "POST" && r.FormValue("wallet") != "" {
		wallet := r.FormValue("wallet")
		// Crear la sesión basada en la billetera criptográfica
		http.SetCookie(w, &http.Cookie{Name: "sesion", Value: wallet, Path: "/"})
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	tmpl, _ := template.ParseFiles("login.html")
	tmpl.Execute(w, nil)
}

func main() {
	http.HandleFunc("/img-maestro", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_MAESTRO) })
	http.HandleFunc("/img-astronauta", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_ASTRONAUTA) })
	http.HandleFunc("/img-quetzal", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_QUETZAL) })
	http.HandleFunc("/img-androide", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_ANDROIDE) })
	http.HandleFunc("/img-supremo", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_SUPREMO) })
	http.HandleFunc("/img-chinampero", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_CHINAMPERO) })
	
	http.HandleFunc("/api/confirm-nft", manejadorConfirmNFT)
	http.HandleFunc("/api/confirm-vote", manejadorConfirmVote)
	
	http.HandleFunc("/login", manejadorLogin)
	http.HandleFunc("/market", manejadorMarket)
	http.HandleFunc("/", manejadorPrincipal)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("--- SERVIDOR WEB3 MULTICUENTA ACTIVADO en puerto %s ---\n", port)
	fmt.Println("Seguridad: Basada en Firmas de MetaMask")
	fmt.Printf("Entra a: http://localhost:%s o tu IP pública/dominio\n", port)
	http.ListenAndServe(":"+port, nil)
}
