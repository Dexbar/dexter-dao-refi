const hre = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  const recipient = process.env.RECIPIENT || process.argv[2];
  const amountStr = process.env.AMOUNT || process.argv[3]; // e.g. "150"
  
  if (!recipient || !amountStr) {
    console.error("Error: Falta dirección del destinatario o cantidad de tokens.");
    console.error("Uso: RECIPIENT=0x... AMOUNT=100 npx hardhat run scripts/mint_tokens.js --network <red>");
    process.exit(1);
  }

  // Cargar las direcciones desplegadas
  const deployedAddressesPath = path.join(__dirname, "../deployed_addresses.json");
  if (!fs.existsSync(deployedAddressesPath)) {
    console.error("Error: deployed_addresses.json no existe. Primero despliega los contratos.");
    process.exit(1);
  }
  
  const deployed = JSON.parse(fs.readFileSync(deployedAddressesPath, "utf8"));
  const tokenAddress = deployed.DexterDAO;
  if (!tokenAddress) {
    console.error("Error: DexterDAO address no especificado en deployed_addresses.json.");
    process.exit(1);
  }

  // Obtener el firmante (deployer/owner)
  const [deployer] = await hre.ethers.getSigners();
  if (!deployer) {
    console.error("Error: No se pudo obtener el signer del deployer. Verifica tu .env.");
    process.exit(1);
  }

  // Conectar al contrato DexterDAO
  const DexterDAO = await hre.ethers.getContractAt("DexterDAO", tokenAddress, deployer);
  
  // Convertir la cantidad a Wei (18 decimales)
  const amount = hre.ethers.parseUnits(amountStr, 18);

  console.log(`[Mint Script] Minteando ${amountStr} DXT a ${recipient} en ${hre.network.name}...`);
  
  const tx = await DexterDAO.mint(recipient, amount);
  console.log(`[Mint Script] Transacción enviada: ${tx.hash}. Esperando confirmación...`);
  
  await tx.wait();
  console.log(`MINT_SUCCESS:${tx.hash}`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error("Error en scripts/mint_tokens.js:", error);
    process.exit(1);
  });
