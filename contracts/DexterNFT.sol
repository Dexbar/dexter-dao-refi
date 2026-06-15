// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract DexterNFT is ERC721, Ownable {
    IERC20 public dxtToken;
    uint256 public nextTokenId;
    string private _baseTokenURI;

    // Estructura para definir los detalles de cada tipo de NFT
    struct NFTType {
        string name;
        uint256 price;       // en tokens DXT (con 18 decimales)
        uint256 maxSupply;
        uint256 currentSupply;
    }

    // Mapeo de ID de tipo -> Estructura de tipo
    mapping(uint256 => NFTType) public nftTypes;
    
    // Mapeo de ID de token -> tipo de NFT
    mapping(uint256 => uint256) public tokenTypes;

    event NFTMinted(address indexed buyer, uint256 indexed tokenId, uint256 indexed nftType);

    constructor(address _dxtTokenAddress) ERC721("Dexter Ajolote NFT", "DXNFT") Ownable(msg.sender) {
        dxtToken = IERC20(_dxtTokenAddress);
        _baseTokenURI = "http://localhost:8080/api/nft/metadata/";
        
        // Inicializar los tipos de Ajolotes NFT según las especificaciones
        // 0: Maestro (ajolote), Price: 200 DXT, Max Supply: 500
        nftTypes[0] = NFTType("Maestro #001", 200 * 10**18, 500, 0);
        // 1: Astronauta (luna), Price: 500 DXT, Max Supply: 100
        nftTypes[1] = NFTType("Astronauta", 500 * 10**18, 100, 0);
        // 2: Dios Quetzal (quetzal), Price: 1000 DXT, Max Supply: 25
        nftTypes[2] = NFTType("Dios Quetzal", 1000 * 10**18, 25, 0);
        // 3: Cyber Androide (androide), Price: 2000 DXT, Max Supply: 5
        nftTypes[3] = NFTType("Cyber Androide", 2000 * 10**18, 5, 0);
        // 4: Ajolote Supremo (supremo), Price: 5000 DXT, Max Supply: 1
        nftTypes[4] = NFTType("Ajolote Supremo", 5000 * 10**18, 1, 0);
        // 5: Ajolote Chinampero (chinampero), Price: 1500 DXT, Max Supply: 200
        nftTypes[5] = NFTType("Ajolote Chinampero", 1500 * 10**18, 200, 0);
        // 6: Ajolote Guardian (guardian), Price: 3000 DXT, Max Supply: 40
        nftTypes[6] = NFTType("Ajolote Guardian", 3000 * 10**18, 40, 0);
        
        // Fútbol Mundialista Edition (Max Supply: 20 each)
        // 7: Ajolote Tricolor (futbol_mex), Price: 1000 DXT, Max Supply: 20
        nftTypes[7] = NFTType("Ajolote Tricolor", 1000 * 10**18, 20, 0);
        // 8: Ajolote Canarinho (futbol_bra), Price: 1000 DXT, Max Supply: 20
        nftTypes[8] = NFTType("Ajolote Canarinho", 1000 * 10**18, 20, 0);
        // 9: Ajolote Albiceleste (futbol_arg), Price: 1000 DXT, Max Supply: 20
        nftTypes[9] = NFTType("Ajolote Albiceleste", 1000 * 10**18, 20, 0);
        // 10: Ajolote Kaiser (futbol_ger), Price: 1000 DXT, Max Supply: 20
        nftTypes[10] = NFTType("Ajolote Kaiser", 1000 * 10**18, 20, 0);
        // 11: Ajolote Furia Roja (futbol_esp), Price: 1000 DXT, Max Supply: 20
        nftTypes[11] = NFTType("Ajolote Furia Roja", 1000 * 10**18, 20, 0);
    }

    // Función para comprar / acuñar un NFT pagando con tokens DXT ERC-20
    function mintNFT(uint256 _nftType) external returns (uint256) {
        require(_nftType <= 11, "Tipo de NFT invalido");
        NFTType storage nft = nftTypes[_nftType];
        require(nft.currentSupply < nft.maxSupply, "Suministro maximo alcanzado para este tipo");

        uint256 price = nft.price;
        
        // Cobrar los tokens DXT al comprador (debe haber aprobado el gasto previamente)
        require(dxtToken.transferFrom(msg.sender, address(this), price), "Transferencia de DXT fallida");

        uint256 tokenId = nextTokenId;
        nextTokenId++;

        nft.currentSupply++;
        tokenTypes[tokenId] = _nftType;

        _safeMint(msg.sender, tokenId);

        emit NFTMinted(msg.sender, tokenId, _nftType);

        return tokenId;
    }

    // Función para obtener la información de un tipo de NFT
    function getNFTTypeInfo(uint256 _nftType) external view returns (
        string memory name,
        uint256 price,
        uint256 maxSupply,
        uint256 currentSupply
    ) {
        require(_nftType <= 11, "Tipo de NFT invalido");
        NFTType memory nft = nftTypes[_nftType];
        return (nft.name, nft.price, nft.maxSupply, nft.currentSupply);
    }

    // Permite al dueño del contrato retirar los tokens DXT recaudados
    function withdrawTokens() external onlyOwner {
        uint256 balance = dxtToken.balanceOf(address(this));
        require(balance > 0, "No hay tokens para retirar");
        require(dxtToken.transfer(owner(), balance), "Retiro fallido");
    }

    // Override _baseURI from OpenZeppelin ERC721
    function _baseURI() internal view virtual override returns (string memory) {
        return _baseTokenURI;
    }

    // Allow owner to set a new base URI
    function setBaseURI(string calldata newBaseURI) external onlyOwner {
        _baseTokenURI = newBaseURI;
    }
}
