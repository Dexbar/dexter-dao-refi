// mint_tokens.js - Sin dependencia de hardhat para funcionar en produccion (Render)
const { ethers } = require("ethers");
const fs = require("fs");
const path = require("path");

// ABI minimo del contrato DexterDAO (solo la funcion mint que necesitamos)
const DEXTER_DAO_ABI = [
  "function mint(address to, uint256 amount) external",
  "function balanceOf(address account) view returns (uint256)",
  "function owner() view returns (address)"
];

async function main() {
  const recipient = process.env.RECIPIENT || process.argv[2];
  const amountStr  = process.env.AMOUNT   || process.argv[3]; // e.g. "150"

  if (!recipient || !amountStr) {
    console.error("Error: Falta direccion del destinatario o cantidad de tokens.");
    process.exit(1);
  }

  // Cargar las direcciones desplegadas
  const deployedPath = path.join(__dirname, "../deployed_addresses.json");
  if (!fs.existsSync(deployedPath)) {
    console.error("Error: deployed_addresses.json no existe.");
    process.exit(1);
  }

  const deployed    = JSON.parse(fs.readFileSync(deployedPath, "utf8"));
  const tokenAddress = deployed.DexterDAO;
  const network      = deployed.network || "sepolia";

  if (!tokenAddress) {
    console.error("Error: DexterDAO address no encontrado en deployed_addresses.json.");
    process.exit(1);
  }

  // Determinar la URL del RPC segun la red
  let rpcUrl;
  if (network === "localhost" || network === "hardhat") {
    rpcUrl = "http://127.0.0.1:8545";
  } else if (network === "sepolia") {
    rpcUrl = process.env.SEPOLIA_RPC_URL || "https://ethereum-sepolia-rpc.publicnode.com";
  } else if (network === "alfajores") {
    rpcUrl = process.env.ALFAJORES_RPC_URL || "https://alfajores-forno.celo-testnet.org";
  } else {
    rpcUrl = process.env.RPC_URL || "https://ethereum-sepolia-rpc.publicnode.com";
  }

  // Clave privada del deployer (owner del contrato)
  const privateKey = process.env.PRIVATE_KEY;
  if (!privateKey) {
    console.error("Error: PRIVATE_KEY no encontrada en variables de entorno.");
    process.exit(1);
  }

  // Conectar al proveedor y firmar
  const provider = new ethers.JsonRpcProvider(rpcUrl);
  const wallet   = new ethers.Wallet(privateKey, provider);
  const contract = new ethers.Contract(tokenAddress, DEXTER_DAO_ABI, wallet);

  // Convertir la cantidad a Wei (18 decimales)
  const amount = ethers.parseUnits(amountStr, 18);

  console.log(`[Mint] Minteando ${amountStr} DXT a ${recipient} en ${network}...`);

  const tx = await contract.mint(recipient, amount);
  console.log(`[Mint] Transaccion enviada: ${tx.hash}. Esperando confirmacion...`);

  await tx.wait();
  console.log(`MINT_SUCCESS:${tx.hash}`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error("Error en mint_tokens.js:", error.message || error);
    process.exit(1);
  });
