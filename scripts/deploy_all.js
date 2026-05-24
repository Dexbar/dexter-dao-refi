const hre = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  console.log("🚀 Desplegando ecosistema completo de Dexter DAO...");

  // --- 1. Desplegar DexterDAO ERC-20 Token ---
  console.log("\n1️⃣ Desplegando ERC-20: DexterDAO Token...");
  const initialSupply = hre.ethers.parseUnits("1000000", 18);
  const DexterDAO = await hre.ethers.getContractFactory("DexterDAO");
  const token = await DexterDAO.deploy(initialSupply);
  await token.waitForDeployment();
  const tokenAddress = await token.getAddress();
  console.log(`✅ DexterDAO Token (DXT) desplegado en: ${tokenAddress}`);

  // --- 2. Desplegar DexterNFT ERC-721 Collection ---
  console.log("\n2️⃣ Desplegando ERC-721: DexterNFT Collection...");
  const DexterNFT = await hre.ethers.getContractFactory("DexterNFT");
  const nft = await DexterNFT.deploy(tokenAddress);
  await nft.waitForDeployment();
  const nftAddress = await nft.getAddress();
  console.log(`✅ DexterNFT (DXNFT) desplegado en: ${nftAddress}`);

  // --- 3. Desplegar DexterGov Governance ---
  console.log("\n3️⃣ Desplegando Gobernanza: DexterGov...");
  const DexterGov = await hre.ethers.getContractFactory("DexterGov");
  const gov = await DexterGov.deploy(tokenAddress);
  await gov.waitForDeployment();
  const govAddress = await gov.getAddress();
  console.log(`✅ DexterGov desplegado en: ${govAddress}`);

  // --- 4. Guardar Direcciones Desplegadas ---
  const addresses = {
    DexterDAO: tokenAddress,
    DexterNFT: nftAddress,
    DexterGov: govAddress,
    deployedAt: new Date().toISOString(),
    network: hre.network.name
  };

  const configPath = path.join(__dirname, "../deployed_addresses.json");
  fs.writeFileSync(configPath, JSON.stringify(addresses, null, 2));
  console.log(`\n💾 Direcciones guardadas en: ${configPath}`);

  console.log("\n🎉 ¡Despliegue completado con éxito!");
  console.log("------------------------------------");
  console.log(`Token address:      ${tokenAddress}`);
  console.log(`NFT address:        ${nftAddress}`);
  console.log(`Gov address:        ${govAddress}`);
  console.log("------------------------------------");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
