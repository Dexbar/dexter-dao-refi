const hre = require("hardhat");
const { ethers } = hre;

async function main() {
  console.log("🚀 Desplegando DexterDAO Token...");

  // 1,000,000 DXT con 18 decimales
  const initialSupply = ethers.parseUnits("1000000", 18);

  const DexterDAO = await ethers.getContractFactory("DexterDAO");
  const token = await DexterDAO.deploy(initialSupply);

  await token.waitForDeployment();

  const address = await token.getAddress();
  console.log(`✅ DexterDAO Token desplegado en: ${address}`);
  console.log(`💰 Supply inicial: 1,000,000 DXT`);
  console.log(`🔗 Ver en Etherscan: https://sepolia.etherscan.io/address/${address}`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
