const hre = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  const deployed = JSON.parse(fs.readFileSync("deployed_addresses.json", "utf8"));
  const tokenAddress = deployed.DexterDAO;
  const [deployer] = await hre.ethers.getSigners();
  const address = await deployer.getAddress();
  
  const DexterDAO = await hre.ethers.getContractAt("DexterDAO", tokenAddress);
  const balance = await DexterDAO.balanceOf(address);
  console.log(`DXT Balance of ${address}: ${hre.ethers.formatUnits(balance, 18)} DXT`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
