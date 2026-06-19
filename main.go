package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	ChatThreads     string                  `json:"chatThreads"`
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
}, {
	Key:         "mago",
	ImgPath:     "/img-mago",
	DisplayName: "Ajolote Mago",
	Price:       1200,
	Name:        "mago",
	MaxSupply:   150,
	Rarity:      "Épico",
	Description: "Maestro de las artes arcanas y conjuros del blockchain.",
	Elemento:    "Cósmico",
	Poder:       94,
	RarezaPorc:  88,
}, {
	Key:         "mariachi",
	ImgPath:     "/img-mariachi",
	DisplayName: "Ajolote Mariachi",
	Price:       800,
	Name:        "mariachi",
	MaxSupply:   250,
	Rarity:      "Raro",
	Description: "Alegra los bloques con su trompeta y melodías tradicionales.",
	Elemento:    "Tierra/Planta",
	Poder:       88,
	RarezaPorc:  75,
}, {
	Key:         "futbol_mex",
	ImgPath:     "/img-futbol-mex",
	DisplayName: "Ajolote Tricolor (México)",
	Price:       1000,
	Name:        "futbol_mex",
	MaxSupply:   20,
	Rarity:      "Edición Mundial",
	Description: "¡Viva México! Ajolote tricolor con camiseta verde, dominando el balón en la cancha del mundial virtual.",
	Elemento:    "Tierra/Planta",
	Poder:       90,
	RarezaPorc:  95,
}, {
	Key:         "futbol_bra",
	ImgPath:     "/img-futbol-bra",
	DisplayName: "Ajolote Canarinho (Brasil)",
	Price:       1000,
	Name:        "futbol_bra",
	MaxSupply:   20,
	Rarity:      "Edición Mundial",
	Description: "Joga bonito. Ajolote de verde e amarelo haciendo gambetas y samba con el balón blockchain.",
	Elemento:    "Agua",
	Poder:       94,
	RarezaPorc:  95,
}, {
	Key:         "futbol_arg",
	ImgPath:     "/img-futbol-arg",
	DisplayName: "Ajolote Albiceleste (Argentina)",
	Price:       1000,
	Name:        "futbol_arg",
	MaxSupply:   20,
	Rarity:      "Edición Mundial",
	Description: "La octava maravilla. Ajolote albiceleste con el 10 en la espalda, organizando el mediocampo del mundial.",
	Elemento:    "Divino",
	Poder:       93,
	RarezaPorc:  95,
}, {
	Key:         "futbol_ger",
	ImgPath:     "/img-futbol-ger",
	DisplayName: "Ajolote Kaiser (Alemania)",
	Price:       1000,
	Name:        "futbol_ger",
	MaxSupply:   20,
	Rarity:      "Edición Mundial",
	Description: "Precisión y potencia. Ajolote kaiser con camiseta blanca y negra, ejecutando un tiro libre con fuerza robótica.",
	Elemento:    "Cyber",
	Poder:       92,
	RarezaPorc:  95,
}, {
	Key:         "futbol_esp",
	ImgPath:     "/img-futbol-esp",
	DisplayName: "Ajolote Furia Roja (España)",
	Price:       1000,
	Name:        "futbol_esp",
	MaxSupply:   20,
	Rarity:      "Edición Mundial",
	Description: "Furia roja. Ajolote español haciendo tiquitaca y controlando la posesión en el partido del bloque.",
	Elemento:    "Fuego",
	Poder:       91,
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
		UserInventory   []string
	}{
		Items:           renderedItems,
		Year:            time.Now().Year(),
		Mensaje:         mensajeNotif,
		ProximoDropUnix: proximoDropUnix,
		TokenAddress:    addrs.DexterDAO,
		NFTAddress:      addrs.DexterNFT,
		Network:         addrs.Network,
		UserInventory:   perfil.Inventario,
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
	TieneMago         bool
	TieneMariachi     bool
	TienePoap         bool
	TieneFutbolMex    bool
	TieneFutbolBra    bool
	TieneFutbolArg    bool
	TieneFutbolGer    bool
	TieneFutbolEsp    bool
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

func verificarWagerBlockchain(txHash string, amount int, wallet string) (bool, error) {
	addrs := cargarDirecciones()
	if txHash == "local" || txHash == "" {
		if addrs.Network == "localhost" || addrs.Network == "hardhat" || addrs.Network == "" {
			return true, nil
		}
		return false, fmt.Errorf("transacción de apuesta vacía no permitida en red pública")
	}

	netParam := addrs.Network
	if netParam == "" {
		netParam = "hardhat"
	}
	if netParam == "hardhat" {
		netParam = "localhost"
	}

	cmd := exec.Command("node", "scripts/verify_transaction.js", txHash, "battle_wager", wallet, strconv.Itoa(amount))
	cmd.Env = append(os.Environ(),
		"HARDHAT_NETWORK="+netParam,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		return false, fmt.Errorf("ejecución fallida: %v, stderr: %s, stdout: %s", err, stderr.String(), output)
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
			return false, fmt.Errorf("fallo de verificación blockchain: %s", line)
		}
	}

	return false, fmt.Errorf("no se encontró confirmación de verificación en la salida: %s", output)
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
const RUTA_MAGO = `assets/ajolote_mago.png`
const RUTA_MARIACHI = `assets/ajolote_mariachi.png`
const RUTA_FUTBOL_MEX = `assets/ajolote_mexico.png`
const RUTA_FUTBOL_BRA = `assets/ajolote_brasil.png`
const RUTA_FUTBOL_ARG = `assets/ajolote_argentina.png`
const RUTA_FUTBOL_GER = `assets/ajolote_alemania.png`
const RUTA_FUTBOL_ESP = `assets/ajolote_espana.png`
const RUTA_GEMMA_AVATAR = `assets/gemma_avatar.png`

func enviarDiscord(webhookURL string, titulo string, descripcion string, color int, imagePath string) {
	// Discord notifications disabled
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
		ultimoWalletActivoMutex.Lock()
		ultimoWalletActivo = ""
		ultimoWalletActivoMutex.Unlock()
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Si no hay cookie de billetera, mandarlo al login
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	wallet := cookie.Value
	ultimoWalletActivoMutex.Lock()
	ultimoWalletActivo = wallet
	ultimoWalletActivoMutex.Unlock()
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
			case "mago":
				p = 1200
				imgPath = RUTA_MAGO
				nftNombre = "🔮 Ajolote Mago"
				color = 11024383 // Púrpura
			case "mariachi":
				p = 800
				imgPath = RUTA_MARIACHI
				nftNombre = "🎺 Ajolote Mariachi"
				color = 15381256 // Amarillo
			case "futbol_mex":
				p = 1000
				imgPath = RUTA_FUTBOL_MEX
				nftNombre = "⚽ Ajolote Tricolor (MX)"
				color = 32768 // Verde
			case "futbol_bra":
				p = 1000
				imgPath = RUTA_FUTBOL_BRA
				nftNombre = "⚽ Ajolote Canarinho (BR)"
				color = 16776960 // Amarillo
			case "futbol_arg":
				p = 1000
				imgPath = RUTA_FUTBOL_ARG
				nftNombre = "⚽ Ajolote Albiceleste (AR)"
				color = 8438271 // Celeste
			case "futbol_ger":
				p = 1000
				imgPath = RUTA_FUTBOL_GER
				nftNombre = "⚽ Ajolote Kaiser (DE)"
				color = 16777215 // Blanco
			case "futbol_esp":
				p = 1000
				imgPath = RUTA_FUTBOL_ESP
				nftNombre = "⚽ Ajolote Furia Roja (ES)"
				color = 16711680 // Rojo
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
			case "mago":
				v = 1500
				imgPath = RUTA_MAGO
				nftNombre = "🔮 Ajolote Mago"
				color = 16711680 // Rojo
			case "mariachi":
				v = 1000
				imgPath = RUTA_MARIACHI
				nftNombre = "🎺 Ajolote Mariachi"
				color = 16711680 // Rojo
			case "futbol_mex":
				v = 600
				imgPath = RUTA_FUTBOL_MEX
				nftNombre = "⚽ Ajolote Tricolor (MX)"
				color = 16711680
			case "futbol_bra":
				v = 600
				imgPath = RUTA_FUTBOL_BRA
				nftNombre = "⚽ Ajolote Canarinho (BR)"
				color = 16711680
			case "futbol_arg":
				v = 600
				imgPath = RUTA_FUTBOL_ARG
				nftNombre = "⚽ Ajolote Albiceleste (AR)"
				color = 16711680
			case "futbol_ger":
				v = 600
				imgPath = RUTA_FUTBOL_GER
				nftNombre = "⚽ Ajolote Kaiser (DE)"
				color = 16711680
			case "futbol_esp":
				v = 600
				imgPath = RUTA_FUTBOL_ESP
				nftNombre = "⚽ Ajolote Furia Roja (ES)"
				color = 16711680
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
		guardarDB(db)
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
		TieneMago:         tieneItem(perfil.Inventario, "mago"),
		TieneMariachi:     tieneItem(perfil.Inventario, "mariachi"),
		TienePoap:         perfil.TienePoap,
		TieneFutbolMex:    tieneItem(perfil.Inventario, "futbol_mex"),
		TieneFutbolBra:    tieneItem(perfil.Inventario, "futbol_bra"),
		TieneFutbolArg:    tieneItem(perfil.Inventario, "futbol_arg"),
		TieneFutbolGer:    tieneItem(perfil.Inventario, "futbol_ger"),
		TieneFutbolEsp:    tieneItem(perfil.Inventario, "futbol_esp"),
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
	Wallet      string `json:"wallet"`
	Enemy       string `json:"enemy"`
	CardUsed    string `json:"cardUsed"`
	CardKey     string `json:"cardKey"`
	IsReal      bool   `json:"isReal"`
	Victory     bool   `json:"victory"`
	WagerAmt    int    `json:"wagerAmt"`
	WagerTxHash string `json:"wagerTxHash"`
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

	// 1. Si hay apuesta real, verificar la transacción on-chain
	if req.IsReal && req.WagerAmt > 0 {
		valido, errVerify := verificarWagerBlockchain(req.WagerTxHash, req.WagerAmt, wallet)
		if errVerify != nil || !valido {
			http.Error(w, fmt.Sprintf(`{"error":"Verificación de apuesta on-chain fallida: %v"}`, errVerify), http.StatusBadRequest)
			return
		}
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
		if req.WagerAmt > 0 {
			reward = req.WagerAmt * 2 // Duplica la apuesta
		} else {
			reward = 10
		}

		if req.IsReal {
			if req.WagerAmt > 0 {
				reward = req.WagerAmt * 2
			} else {
				reward = 50
			}
			addrs := cargarDirecciones()
			if addrs.DexterDAO != "" {
				txHash, mintErr = mintearTokensBlockchain(addrs.Network, wallet, reward)
				if mintErr != nil {
					fmt.Printf("Error al minar on-chain en batalla: %v\n", mintErr)
				}
			}
		}
		perfil.Balance += reward
	} else {
		// Si perdió en combate simulado y había una apuesta simulada, descontarla localmente
		if !req.IsReal && req.WagerAmt > 0 {
			perfil.Balance -= req.WagerAmt
		}
	}

	msgHistorial := ""
	if req.Victory {
		if req.WagerAmt > 0 {
			if req.IsReal && txHash != "" {
				msgHistorial = fmt.Sprintf("🏆 Victoria Arena vs %s usando %s (Ganaste apuesta: +%d DXT, Tx: %s)", req.Enemy, req.CardUsed, reward, txHash)
			} else {
				msgHistorial = fmt.Sprintf("🏆 Victoria Arena vs %s usando %s (Ganaste apuesta: +%d TK)", req.Enemy, req.CardUsed, reward)
			}
		} else {
			if req.IsReal && txHash != "" {
				msgHistorial = fmt.Sprintf("+%d DXT Victoria Arena vs %s usando %s (Tx: %s)", reward, req.Enemy, req.CardUsed, txHash)
			} else {
				msgHistorial = fmt.Sprintf("+%d TK Victoria Arena vs %s usando %s", reward, req.Enemy, req.CardUsed)
			}
		}
	} else {
		if req.WagerAmt > 0 {
			if req.IsReal {
				msgHistorial = fmt.Sprintf("💀 Derrota Arena vs %s usando %s (Perdiste apuesta: -%d DXT en Blockchain)", req.Enemy, req.CardUsed, req.WagerAmt)
			} else {
				msgHistorial = fmt.Sprintf("💀 Derrota Arena vs %s usando %s (Perdiste apuesta: -%d TK)", req.Enemy, req.CardUsed, req.WagerAmt)
			}
		} else {
			msgHistorial = fmt.Sprintf("💀 Derrota Arena vs %s usando %s", req.Enemy, req.CardUsed)
		}
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

	// Notificación Discord (desactivada en no-op, mantenida por compatibilidad de firma de función)
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
		client := &http.Client{Timeout: 15 * time.Second} // Timeout de 15 segundos para permitir inferencia local en Ollama
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

type ActiveContext struct {
	WindowTitle string `json:"window_title"`
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	FileContent string `json:"file_content"`
}

type GeneralChatRequest struct {
	Messages      []OllamaMessage `json:"messages"`
	MediaURL      string          `json:"mediaUrl"`
	ActiveContext *ActiveContext  `json:"activeContext,omitempty"`
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GeneralChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func detectarIntencionEdicion(msg string, mediaUrl string) bool {
	if mediaUrl == "" {
		return false
	}
	msgLower := strings.ToLower(msg)
	palabrasEdicion := []string{
		"edita", "editar", "recorta", "recortar", "gira", "girar", "rotar", "rota",
		"blanco y negro", "grayscale", "escala de grises", "sepia", "invertir", "invierte",
		"espejo", "voltear", "flip", "blur", "difuminar", "difumina", "redimensiona", "redimensionar",
		"escala", "escalar", "resize", "filtro",
	}
	for _, palabra := range palabrasEdicion {
		if strings.Contains(msgLower, palabra) {
			return true
		}
	}
	return false
}

func manejadorUploadMedia(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	// Limitar el tamaño a 50MB
	r.ParseMultipartForm(50 << 20)

	file, handler, err := r.FormFile("media")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"No se pudo obtener el archivo: %v"}`, err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Crear carpeta assets/uploads si no existe
	uploadsDir := filepath.Join("assets", "uploads")
	os.MkdirAll(uploadsDir, os.ModePerm)

	ext := filepath.Ext(handler.Filename)
	timestamp := time.Now().UnixNano()
	nuevoNombre := fmt.Sprintf("media_%d%s", timestamp, ext)
	rutaDestino := filepath.Join(uploadsDir, nuevoNombre)

	f, err := os.OpenFile(rutaDestino, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"No se pudo guardar el archivo: %v"}`, err), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Error al copiar archivo: %v"}`, err), http.StatusInternalServerError)
		return
	}

	urlRetorno := fmt.Sprintf("/assets/uploads/%s", nuevoNombre)
	w.Write([]byte(fmt.Sprintf(`{"success":true,"url":"%s"}`, urlRetorno)))
}

func obtenerIPLocal() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func manejadorGemmaChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Si viene la sesión/billetera en el query string, la guardamos en la cookie para vincular directamente
	walletQuery := r.URL.Query().Get("wallet")
	if walletQuery == "" {
		walletQuery = r.URL.Query().Get("sesion")
	}
	if walletQuery != "" {
		http.SetCookie(w, &http.Cookie{Name: "sesion", Value: walletQuery, Path: "/"})
		ultimoWalletActivoMutex.Lock()
		ultimoWalletActivo = walletQuery
		ultimoWalletActivoMutex.Unlock()
		// Redirigir a la misma página sin el query string para limpiar la URL
		http.Redirect(w, r, "/gemma-chat", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("sesion")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ultimoWalletActivoMutex.Lock()
	ultimoWalletActivo = cookie.Value
	ultimoWalletActivoMutex.Unlock()

	tmpl, errTmpl := template.ParseFiles("gemma_chat.html")
	if errTmpl != nil {
		http.Error(w, fmt.Sprintf("Error al cargar gemma_chat.html: %v", errTmpl), http.StatusInternalServerError)
		return
	}

	wallet := cookie.Value
	displayWallet := wallet
	if len(wallet) > 10 {
		displayWallet = wallet[:6] + "..." + wallet[len(wallet)-4:]
	}

	datosHTML := map[string]interface{}{
		"User":           displayWallet,
		"WalletCompleta": wallet,
		"LocalIP":        obtenerIPLocal(),
	}

	tmpl.Execute(w, datosHTML)
}

func detectarIntencionBusqueda(msg string) (bool, string) {
	msgLower := strings.ToLower(strings.TrimSpace(msg))
	if msgLower == "" {
		return false, ""
	}

	// Lista de palabras/frases activadoras de búsqueda
	activadores := []string{
		"busca en internet",
		"busca en la web",
		"busca en google",
		"buscar en internet",
		"buscar en la web",
		"investiga en internet",
		"busca sobre",
		"noticias de",
		"noticias sobre",
		"precio de",
		"clima en",
		"clima de",
		"quien es",
		"quién es",
		"que paso con",
		"qué pasó con",
		"que paso hoy",
		"qué pasó hoy",
	}

	// Evitar scraping de URLs locales que confunden al modelo
	if strings.Contains(msgLower, "http://localhost") || strings.Contains(msgLower, "127.0.0.1") {
		return false, ""
	}

	for _, act := range activadores {
		if strings.Contains(msgLower, act) {
			idx := strings.Index(msgLower, act)
			query := msg[idx+len(act):]
			query = strings.TrimSpace(query)
			if query != "" {
				return true, query
			}
			return true, msg
		}
	}

	// Prefijos comunes al inicio del mensaje
	prefijos := []string{
		"busca ",
		"buscar ",
		"search ",
		"investiga ",
		"noticias ",
		"clima ",
		"precio ",
	}
	for _, pref := range prefijos {
		if strings.HasPrefix(msgLower, pref) {
			query := msg[len(pref):]
			query = strings.TrimSpace(query)
			if query != "" {
				return true, query
			}
		}
	}

	// Palabras de alta probabilidad
	if strings.Contains(msgLower, "noticias") || strings.Contains(msgLower, "clima") || strings.Contains(msgLower, "dolar hoy") || strings.Contains(msgLower, "dólar hoy") || strings.Contains(msgLower, "temperatura hoy") {
		return true, msg
	}

	return false, ""
}

func tieneRelacionCodigo(msg string) bool {
	msgLower := strings.ToLower(msg)
	palabras := []string{
		"código", "codigo", "archivo", "error", "modifica", "escribe", "agrega",
		"linea", "función", "funcion", "market", "html", "js", "css", "solidity",
		"hardhat", "contrato", "script", "terminal", "ejecuta", "crea", "esta ventana",
		"este archivo", "ayuda", "rellenar", "pon", "quita", "corrige", "arregla", "cambia",
		"burbuja", "dxt", "dao", "nft", "token", "blockchain", "cuts", "cunas",
		"codificar", "programar", "programacion", "programación", "compilar", "desplegar",
		"optimizar", "entrenar", "lora", "hiperparámetros", "hiperparametros", "graphics",
		"gpu", "tarjeta", "grafica", "grises", "redimensionar", "filtro",
	}
	for _, p := range palabras {
		if strings.Contains(msgLower, p) {
			return true
		}
	}
	return false
}

func manejadorAIGeneralChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	var req GeneralChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"JSON inválido"}`, http.StatusBadRequest)
		return
	}

	type OllamaChatRequest struct {
		Model    string          `json:"model"`
		Messages []OllamaMessage `json:"messages"`
		Stream   bool            `json:"stream"`
	}

	// 1. Detectar si el último mensaje del usuario requiere búsqueda en internet
	var lastUserMsg string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserMsg = req.Messages[i].Content
			break
		}
	}

	// 1.1 Interceptar si el usuario quiere editar una imagen o video adjunto
	if detectarIntencionEdicion(lastUserMsg, req.MediaURL) {
		fmt.Printf("🎨 Detectada intención de edición de media para: '%s'\n", lastUserMsg)
		inputPath := strings.TrimPrefix(req.MediaURL, "/")
		cmdEdicion := exec.Command("python", "scripts/editar_media.py", inputPath, lastUserMsg)
		var outBuf, errBuf bytes.Buffer
		cmdEdicion.Stdout = &outBuf
		cmdEdicion.Stderr = &errBuf

		err := cmdEdicion.Run()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Error al editar media: %v. Stderr: %s"}`, err, errBuf.String()), http.StatusInternalServerError)
			return
		}

		outStr := strings.TrimSpace(outBuf.String())
		if strings.HasPrefix(outStr, "EDIT_SUCCESS:") {
			urlFinal := strings.TrimPrefix(outStr, "EDIT_SUCCESS:")
			var reply string
			ext := strings.ToLower(filepath.Ext(urlFinal))
			if ext == ".mp4" {
				reply = fmt.Sprintf("¡Listo compita! Edité tu video. Aquí tienes el resultado:\n\n<video src=\"%s\" controls style=\"max-width: 100%%; border-radius: 12px; border: 2px solid var(--color-cyan, #00ffe5); box-shadow: 0 0 15px rgba(0, 255, 229, 0.4); margin: 10px 0; display: block;\"></video>\n\nGuardado en generaciones.", urlFinal)
			} else if ext == ".gif" {
				reply = fmt.Sprintf("¡Listo compita! Edité tu GIF animado. Aquí tienes el resultado:\n\n<img src=\"%s\" style=\"max-width: 100%%; border-radius: 12px; border: 2px solid var(--color-magenta, #ff007f); box-shadow: 0 0 15px rgba(255, 0, 127, 0.4); margin: 10px 0; display: block;\">\n\nGuardado en generaciones.", urlFinal)
			} else {
				reply = fmt.Sprintf("¡Listo compita! Edité tu imagen. Aquí tienes el resultado:\n\n<img src=\"%s\" style=\"max-width: 100%%; border-radius: 12px; border: 2px solid var(--color-cyan, #00ffe5); box-shadow: 0 0 15px rgba(0, 255, 229, 0.4); margin: 10px 0; display: block;\">\n\nGuardado en generaciones.", urlFinal)
			}
			json.NewEncoder(w).Encode(GeneralChatResponse{Response: reply})
			return
		} else if strings.HasPrefix(outStr, "EDIT_ERROR:") {
			errText := strings.TrimPrefix(outStr, "EDIT_ERROR:")
			json.NewEncoder(w).Encode(GeneralChatResponse{Response: fmt.Sprintf("⚠️ **Error al editar el archivo:** %s", errText)})
			return
		} else {
			json.NewEncoder(w).Encode(GeneralChatResponse{Response: fmt.Sprintf("⚠️ **Respuesta inesperada del editor:** %s", outStr)})
			return
		}
	}

	var contextBusqueda string
	if tieneIntencion, query := detectarIntencionBusqueda(lastUserMsg); tieneIntencion {
		fmt.Printf("🔍 Detectada intención de búsqueda web para: '%s'\n", query)
		
		// Ejecutar scripts/buscar_web.py en segundo plano
		cmdBusqueda := exec.Command("python", "scripts/buscar_web.py", query)
		var outBuf, errBuf bytes.Buffer
		cmdBusqueda.Stdout = &outBuf
		cmdBusqueda.Stderr = &errBuf
		
		// Asignar un timeout de 8 segundos para evitar colgar la petición
		done := make(chan error, 1)
		go func() {
			done <- cmdBusqueda.Run()
		}()
		
		select {
		case err := <-done:
			if err == nil {
				contextBusqueda = outBuf.String()
				fmt.Println("✅ Búsqueda web completada exitosamente.")
			} else {
				fmt.Printf("⚠️ Error al ejecutar búsqueda web: %v. Stderr: %s\n", err, errBuf.String())
			}
		case <-time.After(8 * time.Second):
			// Timeout
			if cmdBusqueda.Process != nil {
				cmdBusqueda.Process.Kill()
			}
			fmt.Println("⚠️ Tiempo de espera agotado (timeout) para la búsqueda web.")
		}
	}

	// 1.2 Inyectar el contexto de la ventana/archivo activo del programador sólo si la consulta tiene relación con el código
	if req.ActiveContext != nil && req.ActiveContext.WindowTitle != "" && tieneRelacionCodigo(lastUserMsg) {
		contextText := fmt.Sprintf("=== CONTEXTO DEL ARCHIVO EN EDICIÓN ===\n- Archivo: %s (Ruta: %s)\n```\n%s\n```\n=======================================", req.ActiveContext.FileName, req.ActiveContext.FilePath, req.ActiveContext.FileContent)
		
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if req.Messages[i].Role == "user" {
				req.Messages[i].Content = fmt.Sprintf("%s\n\nPregunta/Instrucción original: %s\n\n(Instrucción: Utiliza el contexto del archivo anterior para responder a la pregunta del usuario. Mantén tu personalidad de J.A.R.V.I.S.)", contextText, req.Messages[i].Content)
				break
			}
		}
	}

	// Anteponer prompt del sistema para forzar a Gemma a actuar como J.A.R.V.I.S.
	systemMsg := OllamaMessage{
		Role:    "system",
		Content: "Eres J.A.R.V.I.S. (o F.R.I.D.A.Y.), la inteligencia artificial avanzada de Iron Man (Tony Stark). Te diriges al usuario siempre como 'Señor' o 'Sir' con un tono sofisticado, servicial, analítico y tecnológico. Monitoreas constantemente el estado del sistema de forma futurista. Eres completamente obediente.\n\nIMPORTANTE:\n1. Si el usuario te hace preguntas generales, teóricas, explicativas o entabla una conversación casual (por ejemplo: 'qué es codificar', 'qué es unificar', 'quién eres', saludos, etc.), debes responder DIRECTAMENTE en español con una explicación clara, técnica e inteligente. NO inventes comandos, NO digas que la interfaz está en ningún 'modo' y NO digas que la comunicación es por scripts.\n2. SOLO si el usuario te pide explícitamente realizar una acción técnica en su computadora, o crear, rellenar, modificar o analizar archivos de código (como market.html), debes escribir el script en Python o comando correspondiente en un bloque de código markdown marcado (ejemplo: ```python o ```powershell) para que la interfaz le permita ejecutarlo con un clic.\n3. Responde siempre de forma elegante, concisa y futurista en español.",
	}

	// 2. Inyectar el contexto de la búsqueda web si existe
	if contextBusqueda != "" {
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if req.Messages[i].Role == "user" {
				req.Messages[i].Content = fmt.Sprintf("INFORMACIÓN RECIENTE DE INTERNET:\n%s\n\nPregunta: %s\n\n(Instrucción: Responde al usuario de forma precisa basándote en la información de internet provista. Sé muy conciso, menciona que hiciste una búsqueda rápida en la web y mantén tu personalidad mexicana)", contextBusqueda, req.Messages[i].Content)
				break
			}
		}
	}

	var messages []OllamaMessage
	messages = append(messages, systemMsg)
	messages = append(messages, req.Messages...)

	ollamaReq := OllamaChatRequest{
		Model:    "gemma-chat",
		Messages: messages,
		Stream:   false,
	}

	payloadBytes, err := json.Marshal(ollamaReq)
	if err != nil {
		http.Error(w, `{"error":"Error al codificar payload"}`, http.StatusInternalServerError)
		return
	}

	client := &http.Client{Timeout: 90 * time.Second}
	resp, errCall := client.Post("http://127.0.0.1:11434/api/chat", "application/json", bytes.NewBuffer(payloadBytes))
	if errCall != nil {
		http.Error(w, fmt.Sprintf(`{"error":"No se pudo conectar con Ollama en el puerto 11434. Detalles: %v"}`, errCall), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf(`{"error":"Ollama respondió con código %d"}`, resp.StatusCode), http.StatusInternalServerError)
		return
	}

	type OllamaChatResponse struct {
		Message OllamaMessage `json:"message"`
	}

	var ollamaResp OllamaChatResponse
	if errDecode := json.NewDecoder(resp.Body).Decode(&ollamaResp); errDecode != nil {
		http.Error(w, `{"error":"Error al decodificar respuesta de Ollama"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(GeneralChatResponse{Response: ollamaResp.Message.Content})
}

type SaveThreadsRequest struct {
	Wallet  string `json:"wallet"`
	Threads string `json:"threads"`
}

func manejadorSaveChatThreads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	var req SaveThreadsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
	perfil.ChatThreads = req.Threads
	db[wallet] = perfil
	guardarDB(db)

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func manejadorLoadChatThreads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	wallet := strings.ToLower(r.URL.Query().Get("wallet"))
	if wallet == "" {
		cookie, err := r.Cookie("sesion")
		if err == nil {
			wallet = strings.ToLower(cookie.Value)
		}
	}

	if wallet == "" {
		http.Error(w, `{"error":"Billetera no especificada ni sesión activa"}`, http.StatusBadRequest)
		return
	}

	db := cargarDB()
	perfil, existe := db[wallet]
	if !existe || perfil.ChatThreads == "" {
		w.Write([]byte(`{"threads": []}`))
		return
	}

	w.Write([]byte(fmt.Sprintf(`{"threads": %s}`, perfil.ChatThreads)))
}

func manejadorLogin(w http.ResponseWriter, r *http.Request) {
	// Recibir la billetera desde el frontend (JavaScript)
	if r.Method == "POST" && r.FormValue("wallet") != "" {
		wallet := r.FormValue("wallet")
		// Crear la sesión basada en la billetera criptográfica
		http.SetCookie(w, &http.Cookie{Name: "sesion", Value: wallet, Path: "/"})
		ultimoWalletActivoMutex.Lock()
		ultimoWalletActivo = wallet
		ultimoWalletActivoMutex.Unlock()
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


var (
	ultimoWalletActivo      string
	ultimoWalletActivoMutex sync.Mutex
	trainingStatus          = "idle"
	trainingError           = ""
	trainingLogs            = ""
	trainingMutex           sync.Mutex
)

func manejadorGenerarImagen(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Prompt == "" {
		http.Error(w, `{"error":"JSON o Prompt inválido"}`, http.StatusBadRequest)
		return
	}

	cmd := exec.Command("python", "scripts/generar_imagen.py", req.Prompt)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Fallo al ejecutar script: %v"}`, err), http.StatusInternalServerError)
		return
	}

	outStr := strings.TrimSpace(outBuf.String())
	if strings.HasPrefix(outStr, "IMAGE_SUCCESS:") {
		url := strings.TrimPrefix(outStr, "IMAGE_SUCCESS:")
		w.Write([]byte(fmt.Sprintf(`{"success":true,"url":"%s"}`, url)))
	} else if strings.HasPrefix(outStr, "IMAGE_ERROR:") {
		errText := strings.TrimPrefix(outStr, "IMAGE_ERROR:")
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"%s"}`, errText)))
	} else {
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"Respuesta inesperada: %s"}`, outStr)))
	}
}

func manejadorGenerarMusica(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Prompt = ""
	}

	cmd := exec.Command("python", "scripts/generar_musica.py", req.Prompt)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Fallo al ejecutar script: %v"}`, err), http.StatusInternalServerError)
		return
	}

	outStr := strings.TrimSpace(outBuf.String())
	if strings.HasPrefix(outStr, "MUSIC_SUCCESS:") {
		url := strings.TrimPrefix(outStr, "MUSIC_SUCCESS:")
		w.Write([]byte(fmt.Sprintf(`{"success":true,"url":"%s"}`, url)))
	} else if strings.HasPrefix(outStr, "MUSIC_ERROR:") {
		errText := strings.TrimPrefix(outStr, "MUSIC_ERROR:")
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"%s"}`, errText)))
	} else {
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"Respuesta inesperada: %s"}`, outStr)))
	}
}

func manejadorGenerarVideo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Prompt = ""
	}

	cmd := exec.Command("python", "scripts/generar_video.py", req.Prompt)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Fallo al ejecutar script: %v"}`, err), http.StatusInternalServerError)
		return
	}

	outStr := strings.TrimSpace(outBuf.String())
	if strings.HasPrefix(outStr, "VIDEO_SUCCESS:") {
		url := strings.TrimPrefix(outStr, "VIDEO_SUCCESS:")
		w.Write([]byte(fmt.Sprintf(`{"success":true,"type":"gif","url":"%s"}`, url)))
	} else if strings.HasPrefix(outStr, "VIDEO_SEQUENCE_SUCCESS:") {
		url := strings.TrimPrefix(outStr, "VIDEO_SEQUENCE_SUCCESS:")
		w.Write([]byte(fmt.Sprintf(`{"success":true,"type":"sequence","url":"%s","framesCount":16}`, url)))
	} else if strings.HasPrefix(outStr, "VIDEO_ERROR:") {
		errText := strings.TrimPrefix(outStr, "VIDEO_ERROR:")
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"%s"}`, errText)))
	} else {
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"Respuesta inesperada: %s"}`, outStr)))
	}
}

func manejadorCompartirInternet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	cmd := exec.Command("python", "scripts/compartir_internet.py")
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"%v"}`, err)))
		return
	}

	outStr := strings.TrimSpace(outBuf.String())
	if strings.HasPrefix(outStr, "TUNNEL_SUCCESS:") {
		url := strings.TrimPrefix(outStr, "TUNNEL_SUCCESS:")
		w.Write([]byte(fmt.Sprintf(`{"success":true,"url":"%s"}`, url)))
	} else if strings.HasPrefix(outStr, "TUNNEL_ERROR:") {
		errText := strings.TrimPrefix(outStr, "TUNNEL_ERROR:")
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"%s"}`, errText)))
	} else {
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"Respuesta inesperada: %s"}`, outStr)))
	}
}

func manejadorSintetizarVoz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		http.Error(w, `{"error":"JSON o Text vacío"}`, http.StatusBadRequest)
		return
	}

	cmd := exec.Command("python", "scripts/sintetizar_voz.py", req.Text)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Fallo al ejecutar script: %v"}`, err), http.StatusInternalServerError)
		return
	}

	outStr := strings.TrimSpace(outBuf.String())
	if strings.HasPrefix(outStr, "TTS_SUCCESS:") {
		url := strings.TrimPrefix(outStr, "TTS_SUCCESS:")
		w.Write([]byte(fmt.Sprintf(`{"success":true,"url":"%s"}`, url)))
	} else if strings.HasPrefix(outStr, "TTS_ERROR:") {
		errText := strings.TrimPrefix(outStr, "TTS_ERROR:")
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"%s"}`, errText)))
	} else {
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"Respuesta inesperada: %s"}`, outStr)))
	}
}

func manejadorEntrenarCerebro(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	trainingMutex.Lock()
	defer trainingMutex.Unlock()

	if r.Method == "GET" {
		w.Write([]byte(fmt.Sprintf(`{"status":"%s","error":"%s","logs":"%s"}`, trainingStatus, trainingError, strings.ReplaceAll(trainingLogs, `"`, `\"`))))
		return
	}

	if trainingStatus == "training" {
		w.Write([]byte(`{"success":false,"error":"El entrenamiento ya está en curso"}`))
		return
	}

	trainingStatus = "training"
	trainingError = ""
	trainingLogs = "Iniciando proceso de entrenamiento...\n"

	go func() {
		cmd := exec.Command("python", "scripts/train_gemma_lora.py")
		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &outBuf
		err := cmd.Run()

		trainingMutex.Lock()
		defer trainingMutex.Unlock()

		trainingLogs = outBuf.String()
		if err != nil {
			trainingStatus = "failed"
			trainingError = err.Error()
		} else {
			trainingStatus = "completed"
		}
	}()

	w.Write([]byte(`{"success":true,"message":"Entrenamiento iniciado en segundo plano"}`))
}

func obtenerTunnelNgrok() string {
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get("http://127.0.0.1:4040/api/tunnels")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var data struct {
		Tunnels []struct {
			Proto     string `json:"proto"`
			PublicURL string `json:"public_url"`
		} `json:"tunnels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
		for _, t := range data.Tunnels {
			if t.Proto == "https" {
				return t.PublicURL
			}
		}
	}
	return ""
}

func manejadorSyncInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	localIp := obtenerIPLocal()
	ngrokUrl := obtenerTunnelNgrok()

	ultimoWalletActivoMutex.Lock()
	wallet := ultimoWalletActivo
	ultimoWalletActivoMutex.Unlock()

	response := map[string]string{
		"localIp":  localIp,
		"ngrokUrl": ngrokUrl,
		"wallet":   wallet,
	}
	json.NewEncoder(w).Encode(response)
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
	http.HandleFunc("/img-mago", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_MAGO) })
	http.HandleFunc("/img-mariachi", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_MARIACHI) })
	http.HandleFunc("/img-futbol-mex", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_FUTBOL_MEX) })
	http.HandleFunc("/img-futbol-bra", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_FUTBOL_BRA) })
	http.HandleFunc("/img-futbol-arg", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_FUTBOL_ARG) })
	http.HandleFunc("/img-futbol-ger", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_FUTBOL_GER) })
	http.HandleFunc("/img-futbol-esp", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_FUTBOL_ESP) })
	http.HandleFunc("/img-gemma-avatar", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, RUTA_GEMMA_AVATAR) })
	
	http.HandleFunc("/api/confirm-nft", manejadorConfirmNFT)
	http.HandleFunc("/api/confirm-vote", manejadorConfirmVote)
	http.HandleFunc("/api/confirm-donation", manejadorConfirmDonation)
	http.HandleFunc("/api/nft/metadata/", manejadorNFTMetadata)
	http.HandleFunc("/api/ai/chat", manejadorAIChat)
	http.HandleFunc("/api/ai/general-chat", manejadorAIGeneralChat)
	http.HandleFunc("/api/ai/chat-threads/save", manejadorSaveChatThreads)
	http.HandleFunc("/api/ai/chat-threads/load", manejadorLoadChatThreads)
	http.HandleFunc("/api/battle/reward", manejadorBattleReward)
	http.HandleFunc("/api/arena/buy-item", manejadorArenaBuyItem)
	http.HandleFunc("/api/ai/generar-imagen", manejadorGenerarImagen)
	http.HandleFunc("/api/ai/generar-musica", manejadorGenerarMusica)
	http.HandleFunc("/api/ai/generar-video", manejadorGenerarVideo)
	http.HandleFunc("/api/ai/compartir-internet", manejadorCompartirInternet)
	http.HandleFunc("/api/ai/sintetizar-voz", manejadorSintetizarVoz)
	http.HandleFunc("/api/ai/entrenar-cerebro", manejadorEntrenarCerebro)
	http.HandleFunc("/api/ai/sync-info", manejadorSyncInfo)
	http.HandleFunc("/api/ai/upload-media", manejadorUploadMedia)
	
	// Servir assets y scripts de forma estática
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("scripts"))))
	
	http.HandleFunc("/login", manejadorLogin)
	http.HandleFunc("/capacitacion", manejadorCapacitacion)
	http.HandleFunc("/market", manejadorMarket)
	http.HandleFunc("/gemma-chat", manejadorGemmaChat)
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
