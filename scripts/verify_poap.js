// verify_poap.js - Verify ownership of the Devconnect POAP on Gnosis Chain
const { ethers } = require("ethers");

async function main() {
  const walletAddress = process.argv[2];
  if (!walletAddress) {
    console.error("Error: Falta la dirección de la billetera.");
    process.exit(1);
  }

  const wallet = walletAddress.toLowerCase();

  // MOCK: Si es la cuenta por defecto de Hardhat, simular que posee el POAP para facilitar pruebas locales
  if (wallet === "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266") {
    console.log("POAP_VERIFY_SUCCESS:true");
    process.exit(0);
  }

  const rpcUrl = "https://rpc.gnosischain.com";
  const poapAddress = "0x22C1f6050E56d2876009903609a2cC3fEf23a412";
  const poapId = 7553706; // Devconnect Buenos Aires 2025

  const poapAbi = [
    "function balanceOf(address account, uint256 id) view returns (uint256)"
  ];

  try {
    const provider = new ethers.JsonRpcProvider(rpcUrl);
    const contract = new ethers.Contract(poapAddress, poapAbi, provider);
    
    const balance = await contract.balanceOf(wallet, poapId);
    const ownsPoap = balance > 0n;
    
    console.log(`POAP_VERIFY_SUCCESS:${ownsPoap}`);
  } catch (error) {
    // Si falla la conexión a Gnosis (ej. sin internet), imprimir fallback como falso
    console.warn("Advertencia: No se pudo conectar a Gnosis Chain, usando fallback offline.", error.message);
    console.log("POAP_VERIFY_SUCCESS:false");
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
