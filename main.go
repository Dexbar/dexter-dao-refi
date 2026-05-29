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

type FighterStats struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
}

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
	TienePoap       bool
	FighterStats    map[string]FighterStats `json:"fighterStats"`
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
}, {
	Key:         "guardian",
	ImgPath:     "/img-guardian",
	DisplayName: "Ajolote Guardián",
	Price:       3000,
	Name:        "guardian",
	MaxSupply:   40,
	Rarity:      "Legendario",
	Description: "Forjado en la red Celo con fuego digital y capacidades de protección on-chain.",
	Elemento:    "Fuego",
	Poder:       97,
	RarezaPorc:  98,
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

type DonationProposal struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Recipient    string `json:"recipient"`
	Goal         int    `json:"goal"`
	AmountRaised int    `json:"amountRaised"`
	Completed    bool   `json:"completed"`
	Percent      int    `json:"percent"`
}

type PageData struct {
	User              string
	WalletCompleta    string
	Balance           int
	ValorPesos        int
	Mensaje           string
	Historial         []string
	VotosRobot        int
	VotosPago         int
	TieneAjolote      bool
	TieneLuna         bool
	TieneQuetzal      bool
	TieneAndroide     bool
	TieneSupremo      bool
	TieneChinampero   bool
	TieneGuardian     bool
	TienePoap         bool
	DiscordWebhook    string
	TokenAddress      string
	NFTAddress        string
	GovAddress        string
	CrowdfundAddress  string
	ContractAddress   string
	UltimoReclamo     int64
	TiempoRestante    int64
	Network           string
	DonacionesLocales []DonationProposal
	FighterStatsJSON  string
}

type DeployedAddresses struct {
	DexterDAO       string `json:"DexterDAO"`
	DexterNFT       string `json:"DexterNFT"`
	DexterGov       string `json:"DexterGov"`
	DexterCrowdfund string `json:"DexterCrowdfund"`
	Network         string `json:"network"`
}

func cargarDonacionesLocales() []DonationProposal {
	datos, err := os.ReadFile("db_donations.json")
	var donaciones []DonationProposal
	if err == nil {
		json.Unmarshal(datos, &donaciones)
		for i := range donaciones {
			if donaciones[i].Goal > 0 {
				donaciones[i].Percent = donaciones[i].AmountRaised * 100 / donaciones[i].Goal
				if donaciones[i].Percent > 100 {
					donaciones[i].Percent = 100
				}
			}
		}
	}
	return donaciones
}

func guardarDonacionesLocales(donaciones []DonationProposal) {
	datos, _ := json.MarshalIndent(donaciones, "", "  ")
	os.WriteFile("db_donations.json", datos, 0644)
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

func verificarTransaccionBlockchain(txHash string, txType string, wallet string) (bool, error) {
	addrs := cargarDirecciones()
	// Si la transaccion es "local" o vacia, la aceptamos si la red es local y deseamos permitirlo (como fallback)
	if txHash == "local" || txHash == "" {
		if addrs.Network == "localhost" || addrs.Network == "hardhat" || addrs.Network == "" {
			return true, nil
		}
		return false, fmt.Errorf("transaccion vacia o de fallback no permitida en red publica")
	}

	netParam := addrs.Network
	if netParam == "" {
		netParam = "hardhat"
	}
	if netParam == "hardhat" {
		netParam = "localhost"
	}

	cmd := exec.Command("node", "scripts/verify_transaction.js", txHash, txType, wallet)
	cmd.Env = append(os.Environ(),
		"HARDHAT_NETWORK="+netParam,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		return false, fmt.Errorf("ejecucion fallida: %v, stderr: %s, stdout: %s", err, stderr.String(), output)
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "VERIFY_SUCCESS:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return parts[1] == "true", nil
			}
		}
		if strings.HasPrefix(line, "VERIFY_FAILURE:") {
			return false, fmt.Errorf("fallo de verificacion blockchain: %s", line)
		}
	}

	return false, fmt.Errorf("no se encontro confirmacion de verificacion en la salida: %s", output)
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

type OpenSeaAttribute struct {
	TraitType string      `json:"trait_type"`
	Value     interface{} `json:"value"`
}

type OpenSeaMetadata struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Image       string             `json:"image"`
	ExternalURL string             `json:"external_url"`
	Attributes  []OpenSeaAttribute `json:"attributes"`
}


const RUTA_MAESTRO = `assets/ajolote_maestro.png`
const RUTA_ASTRONAUTA = `assets/ajolote_astronauta.png`
const RUTA_QUETZAL = `assets/ajolote_quetzal.png`
const RUTA_ANDROIDE = `assets/ajolote_androide.png`
const RUTA_SUPREMO = `assets/ajolote_supremo.png`
const RUTA_CHINAMPERO = `assets/ajolote_chinampero.png`
const RUTA_DEVCONNECT = `assets/ajolote_devconnect.png`
const RUTA_GUARDIAN = `assets/ajolote_guardian.png`

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

	// Verificar POAP en segundo plano si aún no está validado
	if !perfil.TienePoap {
		cmd := exec.Command("node", "scripts/verify_poap.js", wallet)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		if err := cmd.Run(); err == nil {
			output := strings.TrimSpace(stdout.String())
			if strings.Contains(output, "POAP_VERIFY_SUCCESS:true") {
				perfil.TienePoap = true
				agregarHistorial(&perfil, "¡POAP Devconnect verificado! Bono de mineria de +5 DXT desbloqueado. 🇦🇷")
				db[wallet] = perfil
				guardarDB(db)
			}
		}
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
				if perfil.TienePoap {
					basePago += 5
					bonoMsg += " (Bono Devconnect +5 DXT! 🎟️)"
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
						switch addrs.Network {
						case "alfajores":
							txLink = fmt.Sprintf("https://celo-alfajores.blockscout.com/tx/%s", txHash)
						case "hardhat", "localhost":
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

		if r.FormValue("proponer_donacion") != "" {
			titulo := r.FormValue("titulo")
			descripcion := r.FormValue("descripcion")
			goal, _ := strconv.Atoi(r.FormValue("goal"))
			recipient := r.FormValue("recipient")
			if titulo != "" && recipient != "" && goal > 0 {
				donaciones := cargarDonacionesLocales()
				newId := len(donaciones)
				donaciones = append(donaciones, DonationProposal{
					ID:           newId,
					Title:        titulo,
					Description:  descripcion,
					Recipient:    recipient,
					Goal:         goal,
					AmountRaised: 0,
					Completed:    false,
				})
				guardarDonacionesLocales(donaciones)
				mensajeNotif = "¡Propuesta local de donación creada!"
				enviarDiscord(perfil.DiscordWebhook, "🗳️ NUEVA PROPUESTA DE DONACIÓN (SIMULADA)", fmt.Sprintf("Billetera: `%s` ha propuesto una iniciativa:\n**%s**\n*Meta:* %d TK\n*Beneficiario:* `%s`", displayWallet, titulo, goal, recipient), 16753920, "")
			}
		}

		if r.FormValue("donar_local") != "" {
			propId, _ := strconv.Atoi(r.FormValue("proposal_id"))
			cantidad, _ := strconv.Atoi(r.FormValue("cantidad"))
			if cantidad > 0 {
				donaciones := cargarDonacionesLocales()
				if propId >= 0 && propId < len(donaciones) {
					prop := &donaciones[propId]
					if perfil.Balance >= cantidad {
						perfil.Balance -= cantidad
						prop.AmountRaised += cantidad
						if prop.AmountRaised >= prop.Goal {
							prop.Completed = true
						}
						agregarHistorial(&perfil, fmt.Sprintf("Donaste %d TK a la propuesta: %s", cantidad, prop.Title))
						guardarDonacionesLocales(donaciones)
						mensajeNotif = fmt.Sprintf("¡Donaste %d TK exitosamente!", cantidad)
						enviarDiscord(perfil.DiscordWebhook, "🎁 DONACIÓN SIMULADA REALIZADA", fmt.Sprintf("Billetera: `%s` donó **%d TK** a la iniciativa:\n**%s**\n*Recaudado:* %d/%d TK", displayWallet, cantidad, prop.Title, prop.AmountRaised, prop.Goal), 16766720, "")
					} else {
						mensajeNotif = "¡Saldo de TK insuficiente para realizar la donación!"
					}
				}
			}
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
			case "guardian":
				p = 3000
				imgPath = RUTA_GUARDIAN
				nftNombre = "🔥 Ajolote Guardián"
				color = 16729856 // Naranja
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
			case "guardian":
				v = 2000
				imgPath = RUTA_GUARDIAN
				nftNombre = "🔥 Ajolote Guardián"
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

	statsJSON, _ := json.Marshal(perfil.FighterStats)
	if perfil.FighterStats == nil {
		statsJSON = []byte("{}")
	}

	datosHTML := PageData{
		User:              displayWallet,
		WalletCompleta:    wallet,
		Balance:           perfil.Balance,
		ValorPesos:        perfil.Balance * 2,
		Mensaje:           mensajeNotif,
		Historial:         perfil.Historial,
		VotosRobot:        perfil.VotosRobot,
		VotosPago:         perfil.VotosPago,
		TieneAjolote:      tieneItem(perfil.Inventario, "ajolote"),
		TieneLuna:         tieneItem(perfil.Inventario, "luna"),
		TieneQuetzal:      tieneItem(perfil.Inventario, "quetzal"),
		TieneAndroide:     tieneItem(perfil.Inventario, "androide"),
		TieneSupremo:      tieneItem(perfil.Inventario, "supremo"),
		TieneChinampero:   tieneItem(perfil.Inventario, "chinampero"),
		TieneGuardian:     tieneItem(perfil.Inventario, "guardian"),
		TienePoap:         perfil.TienePoap,
		DiscordWebhook:    perfil.DiscordWebhook,
		TokenAddress:      addrs.DexterDAO,
		NFTAddress:        addrs.DexterNFT,
		GovAddress:        addrs.DexterGov,
		CrowdfundAddress:  addrs.DexterCrowdfund,
		ContractAddress:   perfil.ContractAddress,
		UltimoReclamo:     perfil.UltimoReclamo,
		TiempoRestante:    tiempoRestante,
		Network:           addrs.Network,
		DonacionesLocales: cargarDonacionesLocales(),
		FighterStatsJSON:  string(statsJSON),
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

	// Verificar la transacción on-chain antes de actualizar la base de datos
	valido, errVerify := verificarTransaccionBlockchain(req.TxHash, "nft_mint", wallet)
	if errVerify != nil || !valido {
		http.Error(w, fmt.Sprintf("Verificacion de transaccion fallida: %v", errVerify), http.StatusBadRequest)
		return
	}

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
	switch req.Key {
	case "supremo":
		color = 16766720 // Oro
	case "chinampero":
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

	// Verificar la transacción on-chain antes de actualizar la base de datos
	valido, errVerify := verificarTransaccionBlockchain(req.TxHash, "proposal_vote", wallet)
	if errVerify != nil || !valido {
		http.Error(w, fmt.Sprintf("Verificacion de transaccion fallida: %v", errVerify), http.StatusBadRequest)
		return
	}

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

type ConfirmDonationRequest struct {
	Wallet     string `json:"wallet"`
	CampaignID int    `json:"campaignId"`
	Amount     string `json:"amount"`
	Type       string `json:"type"`
	TxHash     string `json:"txHash"`
}

func manejadorConfirmDonation(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req ConfirmDonationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	wallet := strings.ToLower(req.Wallet)

	// Verificar la transacción on-chain antes de actualizar la base de datos
	valido, errVerify := verificarTransaccionBlockchain(req.TxHash, "donation", wallet)
	if errVerify != nil || !valido {
		http.Error(w, fmt.Sprintf("Verificacion de transaccion fallida: %v", errVerify), http.StatusBadRequest)
		return
	}

	db := cargarDB()
	perfil, existe := db[wallet]
	if !existe {
		perfil = UserProfile{Balance: 0, Inventario: []string{}, Historial: []string{}}
	}

	msgHistorial := fmt.Sprintf("Donación firmada en Blockchain: %s %s a la Campaña #%d (Tx: %s)", req.Amount, req.Type, req.CampaignID, req.TxHash)
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
		"🎁 DONACIÓN REAL EN BLOCKCHAIN",
		fmt.Sprintf("Billetera: `%s` donó **%s %s** en la blockchain a la Campaña #%d.\n\n**Transacción:** %s", displayWallet, req.Amount, req.Type, req.CampaignID, txLink),
		65280, // Verde
		"",
	)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("confirmado"))
}

type BattleRewardRequest struct {
	Wallet   string `json:"wallet"`
	Enemy    string `json:"enemy"`
	CardUsed string `json:"cardUsed"`
	CardKey  string `json:"cardKey"`
	IsReal   bool   `json:"isReal"`
	Victory  bool   `json:"victory"`
}

type BattleRewardResponse struct {
	Status string `json:"status"`
	Reward int    `json:"reward"`
	TxHash string `json:"txHash,omitempty"`
}

func manejadorBattleReward(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, `{"error":"Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	var req BattleRewardRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error":"JSON inválido"}`, http.StatusBadRequest)
		return
	}

	wallet := strings.ToLower(req.Wallet)
	if wallet == "" {
		http.Error(w, `{"error":"Billetera requerida"}`, http.StatusBadRequest)
		return
	}

	db := cargarDB()
	perfil, existe := db[wallet]
	if !existe {
		perfil = UserProfile{Balance: 0, Inventario: []string{}, Historial: []string{}}
	}

	// Inicializar mapa de estadísticas si es nulo
	if perfil.FighterStats == nil {
		perfil.FighterStats = make(map[string]FighterStats)
	}

	// Actualizar estadísticas del luchador
	stats := perfil.FighterStats[req.CardKey]
	if req.Victory {
		stats.Wins++
	} else {
		stats.Losses++
	}
	perfil.FighterStats[req.CardKey] = stats

	reward := 0
	txHash := ""
	var mintErr error

	if req.Victory {
		reward = 10
		if req.IsReal {
			reward = 50
			addrs := cargarDirecciones()
			if addrs.DexterDAO != "" {
				txHash, mintErr = mintearTokensBlockchain(addrs.Network, wallet, reward)
				if mintErr != nil {
					fmt.Printf("Error al minar on-chain en batalla: %v\n", mintErr)
				}
			}
		}
		perfil.Balance += reward
	}

	msgHistorial := ""
	if req.Victory {
		if req.IsReal && txHash != "" {
			msgHistorial = fmt.Sprintf("+%d DXT Victoria en la Arena vs %s usando %s (Tx: %s)", reward, req.Enemy, req.CardUsed, txHash)
		} else if req.IsReal {
			msgHistorial = fmt.Sprintf("+%d TK Victoria en la Arena vs %s usando %s (Falla on-chain)", reward, req.Enemy, req.CardUsed)
		} else {
			msgHistorial = fmt.Sprintf("+%d TK Victoria en la Arena vs %s usando %s (Modo Práctica)", reward, req.Enemy, req.CardUsed)
		}
	} else {
		msgHistorial = fmt.Sprintf("💀 Derrota en la Arena vs %s usando %s", req.Enemy, req.CardUsed)
	}
	agregarHistorial(&perfil, msgHistorial)

	db[wallet] = perfil
	guardarDB(db)

	displayWallet := wallet
	if len(wallet) > 10 {
		displayWallet = wallet[:6] + "..." + wallet[len(wallet)-4:]
	}

	txLink := ""
	if txHash != "" {
		addrs := cargarDirecciones()
		txLink = fmt.Sprintf("https://sepolia.etherscan.io/tx/%s", txHash)
		if addrs.Network == "alfajores" {
			txLink = fmt.Sprintf("https://celo-alfajores.blockscout.com/tx/%s", txHash)
		} else if addrs.Network == "hardhat" || addrs.Network == "localhost" {
			txLink = "Red Local (Hardhat)"
		}
	}

	// Notificación Discord
	discordMsg := ""
	discordColor := 0x00ff00
	if req.Victory {
		if req.IsReal && txHash != "" {
			discordMsg = fmt.Sprintf("🎮 **¡VICTORIA EN LA ARENA DE BATALLA!**\nBilletera: `%s` ha derrotado a **%s** usando la carta **%s**.\n\n**Recompensa:** +%d DXT (Acuñados on-chain)\n**Transacción:** %s", displayWallet, req.Enemy, req.CardUsed, reward, txLink)
		} else if req.IsReal {
			discordMsg = fmt.Sprintf("🎮 **¡VICTORIA EN LA ARENA DE BATALLA!**\nBilletera: `%s` ha derrotado a **%s** usando la carta **%s**.\n\n**Recompensa:** +%d DXT (Sincronizado localmente por falla de red)", displayWallet, req.Enemy, req.CardUsed, reward)
		} else {
			discordMsg = fmt.Sprintf("🎮 **¡VICTORIA EN LA ARENA DE BATALLA (PRÁCTICA)!**\nBilletera: `%s` ha derrotado a **%s** usando la carta básica **%s**.\n\n**Recompensa:** +%d TK (Solo local/práctica)", displayWallet, req.Enemy, req.CardUsed, reward)
		}
	} else {
		discordColor = 0xff0000
		discordMsg = fmt.Sprintf("💀 **¡DERROTA EN LA ARENA DE BATALLA!**\nBilletera: `%s` fue derrotada por **%s** usando la carta **%s**.", displayWallet, req.Enemy, req.CardUsed)
	}

	enviarDiscord(perfil.DiscordWebhook, "⚔️ AJOLOTE BATTLE ARENA", discordMsg, discordColor, "")

	resp := BattleRewardResponse{
		Status: "success",
		Reward: reward,
		TxHash: txHash,
	}
	json.NewEncoder(w).Encode(resp)
}

type BuyItemRequest struct {
	Wallet   string `json:"wallet"`
	ItemID   string `json:"itemId"`
	ItemName string `json:"itemName"`
	Cost     int    `json:"cost"`
}

type BuyItemResponse struct {
	Status     string `json:"status"`
	NewBalance int    `json:"newBalance"`
}

func manejadorArenaBuyItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, `{"error":"Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	var req BuyItemRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error":"JSON inválido"}`, http.StatusBadRequest)
		return
	}

	wallet := strings.ToLower(req.Wallet)
	if wallet == "" {
		http.Error(w, `{"error":"Billetera requerida"}`, http.StatusBadRequest)
		return
	}

	db := cargarDB()
	perfil, existe := db[wallet]
	if !existe {
		http.Error(w, `{"error":"Billetera no registrada"}`, http.StatusBadRequest)
		return
	}

	if perfil.Balance < req.Cost {
		http.Error(w, `{"error":"Saldo insuficiente de TK"}`, http.StatusBadRequest)
		return
	}

	perfil.Balance -= req.Cost
	msgHistorial := fmt.Sprintf("-%d TK Compra de item: %s 🎒", req.Cost, req.ItemName)
	agregarHistorial(&perfil, msgHistorial)

	db[wallet] = perfil
	guardarDB(db)

	resp := BuyItemResponse{
		Status:     "success",
		NewBalance: perfil.Balance,
	}
	json.NewEncoder(w).Encode(resp)
}

func manejadorNFTMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Extraer tokenId de la URL (ej. /api/nft/metadata/1 -> 1)
	partes := strings.Split(r.URL.Path, "/")
	if len(partes) < 5 {
		http.Error(w, `{"error":"ID de token no especificado"}`, http.StatusBadRequest)
		return
	}
	tokenIdStr := partes[len(partes)-1]
	tokenId, err := strconv.Atoi(tokenIdStr)
	if err != nil {
		http.Error(w, `{"error":"ID de token invalido"}`, http.StatusBadRequest)
		return
	}

	// Consultar el tipo de NFT ejecutando get_nft_info.js
	cmd := exec.Command("node", "scripts/get_nft_info.js", strconv.Itoa(tokenId))
	
	// Cargar variables de entorno necesarias de la red (ej: Sepolia RPC)
	addrs := cargarDirecciones()
	netParam := addrs.Network
	if netParam == "" {
		netParam = "hardhat"
	}
	if netParam == "hardhat" {
		netParam = "localhost"
	}
	cmd.Env = append(os.Environ(),
		"HARDHAT_NETWORK="+netParam,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := stdout.String()

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Fallo al consultar blockchain: %v, stderr: %s"}`, err, stderr.String()), http.StatusInternalServerError)
		return
	}

	// Buscar confirmación en la salida
	var key string
	var owner string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "NFT_INFO_SUCCESS:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 3 {
				key = parts[1]
				owner = parts[2]
			}
		}
	}

	if key == "" {
		http.Error(w, fmt.Sprintf(`{"error":"No se encontro informacion del token en salida: %s"}`, output), http.StatusNotFound)
		return
	}

	// Encontrar la definición del item en marketItems
	var itemFound *MarketItem
	for i := range marketItems {
		if marketItems[i].Key == key {
			itemFound = &marketItems[i]
			break
		}
	}

	if itemFound == nil {
		http.Error(w, `{"error":"Tipo de NFT desconocido"}`, http.StatusNotFound)
		return
	}

	// Construir la URL base dinámica para la imagen
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	imageURL := fmt.Sprintf("%s://%s%s", scheme, host, itemFound.ImgPath)
	externalURL := fmt.Sprintf("%s://%s/market", scheme, host)

	// Crear metadatos compatibles con OpenSea
	descClean := strings.TrimSuffix(itemFound.Description, ".")
	metadata := OpenSeaMetadata{
		Name:        fmt.Sprintf("%s #%d", itemFound.DisplayName, tokenId),
		Description: fmt.Sprintf("%s. Poseído actualmente por %s.", descClean, owner),
		Image:       imageURL,
		ExternalURL: externalURL,
		Attributes: []OpenSeaAttribute{
			{TraitType: "Elemento", Value: itemFound.Elemento},
			{TraitType: "Poder", Value: itemFound.Poder},
			{TraitType: "Rareza", Value: itemFound.Rarity},
			{TraitType: "Porcentaje de Rareza", Value: itemFound.RarezaPorc},
		},
	}

	json.NewEncoder(w).Encode(metadata)
}

type AIChatRequest struct {
	Message string `json:"message"`
}

type GemmaRequest struct {
	Prompt string `json:"prompt"`
}

type GemmaResponse struct {
	Instruction string `json:"instruction"`
	Command     string `json:"command"`
}

type AIChatResponse struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation"`
	Source      string `json:"source"`
}

type TrainItem struct {
	Prompt  string `json:"prompt"`
	Command string `json:"command"`
}

func obtenerExplicacionComando(command string) string {
	cmdClean := strings.ToLower(command)
	if strings.Contains(cmdClean, "compile") {
		return "Este comando compila todos tus contratos inteligentes de Solidity usando Hardhat, generando los artefactos y ABIs correspondientes."
	}
	if strings.Contains(cmdClean, "deploy_all.js") {
		if strings.Contains(cmdClean, "sepolia") {
			return "Despliega todos los contratos inteligentes (DexterDAO, DexterNFT, DexterGov, DexterCrowdfund) a la red de prueba Ethereum Sepolia."
		}
		if strings.Contains(cmdClean, "alfajores") {
			return "Despliega todos los contratos inteligentes del ecosistema a la red de prueba Celo Alfajores."
		}
		return "Despliega todos los contratos inteligentes localmente en tu red de desarrollo de Hardhat simulada."
	}
	if strings.Contains(cmdClean, "check_balance.js") {
		return "Consulta el saldo actual de Ether (ETH) de la billetera local configurada en la red activa."
	}
	if strings.Contains(cmdClean, "check_dxt_balance.js") {
		return "Consulta el saldo del token nativo DXT (Dexter Token) en tu dirección local."
	}
	if strings.Contains(cmdClean, "mint_tokens.js") {
		return "Mintea tokens DXT on-chain hacia la dirección especificada para incrementar el balance en tu cuenta."
	}
	if strings.Contains(cmdClean, "node") {
		return "Inicia un nodo local de Hardhat de forma persistente. Esto crea una blockchain de pruebas en http://127.0.0.1:8545 con 20 cuentas de prueba cargadas con 10000 ETH cada una."
	}
	if strings.Contains(cmdClean, "test_db.go") {
		return "Ejecuta una prueba del backend de Go en 'scratch/test_db.go' para verificar que la persistencia y lectura de 'db.json' funcionen correctamente."
	}
	if strings.Contains(cmdClean, "main.go") {
		return "Ejecuta el servidor web y backend de Go de Dexter DAO localmente en el puerto 8080."
	}
	return "Este comando fue sugerido en base a tu consulta para interactuar con Dexter DAO."
}

func fallbackMatch(message string) (string, string) {
	msgLower := strings.ToLower(strings.TrimSpace(message))

	// Leer terminal_train.jsonl
	datos, err := os.ReadFile("terminal_train.jsonl")
	var bestCommand string
	var bestPrompt string
	maxScore := 0.0

	if err == nil {
		lines := strings.Split(string(datos), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var item TrainItem
			if errJson := json.Unmarshal([]byte(line), &item); errJson == nil {
				promptLower := strings.ToLower(item.Prompt)
				
				// Algoritmo de puntuación simple:
				// 1. Coincidencia exacta
				if msgLower == promptLower {
					return item.Command, "Fallback Local (Coincidencia Exacta)"
				}
				// 2. Substring completo
				if strings.Contains(msgLower, promptLower) || strings.Contains(promptLower, msgLower) {
					score := float64(len(promptLower)) / float64(len(msgLower))
					if score > maxScore {
						maxScore = score
						bestCommand = item.Command
						bestPrompt = item.Prompt
					}
				}
				
				// 3. Intersección de palabras
				promptWords := strings.Fields(promptLower)
				msgWords := strings.Fields(msgLower)
				matches := 0
				for _, pw := range promptWords {
					// Ignorar palabras comunes cortas como "de", "la", "el", "los", "the", "to", "in", "on"
					if len(pw) <= 2 {
						continue
					}
					for _, mw := range msgWords {
						if pw == mw {
							matches++
						}
					}
				}
				if matches > 0 {
					score := float64(matches) / float64(len(promptWords))
					if score > maxScore {
						maxScore = score
						bestCommand = item.Command
						bestPrompt = item.Prompt
					}
				}
			}
		}
	}

	if bestCommand != "" && maxScore >= 0.3 {
		return bestCommand, fmt.Sprintf("Fallback Local (Coincidencia con '%s')", bestPrompt)
	}

	// Reglas de respaldo hardcoded por si no hay coincidencias claras en el jsonl
	if strings.Contains(msgLower, "compile") || strings.Contains(msgLower, "compil") {
		return "npx hardhat compile", "Fallback Local (Regla: Compilar)"
	}
	if strings.Contains(msgLower, "node") || strings.Contains(msgLower, "nodo") || strings.Contains(msgLower, "blockchain local") {
		return "npx hardhat node", "Fallback Local (Regla: Nodo Local)"
	}
	if strings.Contains(msgLower, "deploy") || strings.Contains(msgLower, "desplegar") {
		if strings.Contains(msgLower, "sepolia") {
			return "npx hardhat run scripts/deploy_all.js --network sepolia", "Fallback Local (Regla: Despliegue Sepolia)"
		}
		if strings.Contains(msgLower, "celo") || strings.Contains(msgLower, "alfajores") {
			return "npx hardhat run scripts/deploy_all.js --network alfajores", "Fallback Local (Regla: Despliegue Alfajores)"
		}
		return "npx hardhat run scripts/deploy_all.js --network localhost", "Fallback Local (Regla: Despliegue Local)"
	}
	if strings.Contains(msgLower, "mint") || strings.Contains(msgLower, "mintear") || strings.Contains(msgLower, "minar") || strings.Contains(msgLower, "chamba") {
		// Buscar dirección 0x en el mensaje
		words := strings.Fields(msgLower)
		address := "<dirección_billetera>"
		amount := "<cantidad>"
		for _, w := range words {
			if strings.HasPrefix(w, "0x") && len(w) == 42 {
				address = w
			} else if _, errNum := strconv.Atoi(w); errNum == nil {
				amount = w
			}
		}
		return fmt.Sprintf("node scripts/mint_tokens.js %s %s", address, amount), "Fallback Local (Regla: Mintear DXT)"
	}
	if strings.Contains(msgLower, "balance") || strings.Contains(msgLower, "saldo") {
		if strings.Contains(msgLower, "dxt") || strings.Contains(msgLower, "token") {
			return "node scripts/check_dxt_balance.js", "Fallback Local (Regla: Saldo DXT)"
		}
		return "node scripts/check_balance.js", "Fallback Local (Regla: Saldo ETH)"
	}
	if strings.Contains(msgLower, "go") || strings.Contains(msgLower, "server") || strings.Contains(msgLower, "servidor") || strings.Contains(msgLower, "backend") {
		return "go run main.go", "Fallback Local (Regla: Servidor Go)"
	}
	if strings.Contains(msgLower, "db") || strings.Contains(msgLower, "database") || strings.Contains(msgLower, "base de datos") || strings.Contains(msgLower, "json") {
		return "go run scratch/test_db.go", "Fallback Local (Regla: Base de Datos)"
	}

	return "", ""
}

func manejadorAIChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Permitir CORS básico por seguridad
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, `{"error":"Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	var req AIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		http.Error(w, `{"error":"El mensaje no puede estar vacío"}`, http.StatusBadRequest)
		return
	}

	// 1. Intentar contactar con FastAPI (servir_gemma.py)
	gemmaReq := GemmaRequest{Prompt: req.Message}
	payloadBytes, err := json.Marshal(gemmaReq)
	if err == nil {
		client := &http.Client{Timeout: 2 * time.Second} // Timeout corto de 2 segundos para responder rápido
		resp, errCall := client.Post("http://localhost:8000/predict", "application/json", bytes.NewBuffer(payloadBytes))
		if errCall == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var gemmaResp GemmaResponse
				if errDecode := json.NewDecoder(resp.Body).Decode(&gemmaResp); errDecode == nil && gemmaResp.Command != "" {
					explanation := obtenerExplicacionComando(gemmaResp.Command)
					response := AIChatResponse{
						Command:     gemmaResp.Command,
						Explanation: explanation,
						Source:      "Gemma AI (Servidor de Inferencia)",
					}
					json.NewEncoder(w).Encode(response)
					return
				}
			}
		}
	}

	// 2. Si falló Gemma, usar fallback local en Go
	cmdMatch, source := fallbackMatch(req.Message)
	if cmdMatch != "" {
		explanation := obtenerExplicacionComando(cmdMatch)
		response := AIChatResponse{
			Command:     cmdMatch,
			Explanation: explanation,
			Source:      source,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 3. Fallback genérico si no se entiende
	response := AIChatResponse{
		Command:     "npx hardhat help",
		Explanation: "No pude descifrar el comando específico para tu consulta. Puedes consultar la ayuda general de Hardhat ejecutando este comando.",
		Source:      "Fallback Local (Sin coincidencias)",
	}
	json.NewEncoder(w).Encode(response)
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

func manejadorCapacitacion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.ServeFile(w, r, "capacitacion.html")
}

func cargarEnv() {
	datos, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	lineas := strings.Split(string(datos), "\n")
	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)
		if linea == "" || strings.HasPrefix(linea, "#") {
			continue
		}
		partes := strings.SplitN(linea, "=", 2)
		if len(partes) == 2 {
			clave := strings.TrimSpace(partes[0])
			valor := strings.TrimSpace(partes[1])
			valor = strings.Trim(valor, `"'`)
			os.Setenv(clave, valor)
		}
	}
}

func main() {
	cargarEnv()
	http.HandleFunc("/img-maestro", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_MAESTRO) })
	http.HandleFunc("/img-astronauta", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_ASTRONAUTA) })
	http.HandleFunc("/img-quetzal", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_QUETZAL) })
	http.HandleFunc("/img-androide", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_ANDROIDE) })
	http.HandleFunc("/img-supremo", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_SUPREMO) })
	http.HandleFunc("/img-chinampero", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_CHINAMPERO) })
	http.HandleFunc("/img-devconnect", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_DEVCONNECT) })
	http.HandleFunc("/img-guardian", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_GUARDIAN) })
	
	http.HandleFunc("/api/confirm-nft", manejadorConfirmNFT)
	http.HandleFunc("/api/confirm-vote", manejadorConfirmVote)
	http.HandleFunc("/api/confirm-donation", manejadorConfirmDonation)
	http.HandleFunc("/api/nft/metadata/", manejadorNFTMetadata)
	http.HandleFunc("/api/ai/chat", manejadorAIChat)
	http.HandleFunc("/api/battle/reward", manejadorBattleReward)
	http.HandleFunc("/api/arena/buy-item", manejadorArenaBuyItem)
	
	// Servir scripts de forma estática para la descarga de los scripts de Python
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("scripts"))))
	
	http.HandleFunc("/login", manejadorLogin)
	http.HandleFunc("/capacitacion", manejadorCapacitacion)
	http.HandleFunc("/market", manejadorMarket)
	http.HandleFunc("/", manejadorPrincipal)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Iniciar escuchador de eventos de la blockchain en segundo plano (daemon)
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("🚀 Iniciando daemon blockchain_listener.js en segundo plano...")
		cmdListener := exec.Command("node", "scripts/blockchain_listener.js")
		cmdListener.Stdout = os.Stdout
		cmdListener.Stderr = os.Stderr
		errListener := cmdListener.Run()
		if errListener != nil {
			fmt.Printf("⚠️ El daemon blockchain_listener.js se detuvo con error: %v\n", errListener)
		}
	}()

	fmt.Printf("--- SERVIDOR WEB3 MULTICUENTA ACTIVADO en puerto %s ---\n", port)
	fmt.Println("Seguridad: Basada en Firmas de MetaMask")
	fmt.Printf("Entra a: http://localhost:%s o tu IP pública/dominio\n", port)
	http.ListenAndServe(":"+port, nil)
}
