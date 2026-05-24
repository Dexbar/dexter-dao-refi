// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract DexterGov is Ownable {
    IERC20 public dxtToken;

    struct Proposal {
        uint256 id;
        string description;
        uint256 voteCount;       // Poder de voto total acumulado (balance DXT)
        uint256 votesCountRaw;   // Cantidad de votos individuales (número de personas)
    }

    Proposal[] public proposals;
    
    // Mapeo: ID Propuesta => Dirección => Ha Votado
    mapping(uint256 => mapping(address => bool)) public hasVoted;

    event Voted(address indexed voter, uint256 indexed proposalId, uint256 weight);
    event ProposalAdded(uint256 indexed proposalId, string description);

    constructor(address _dxtTokenAddress) Ownable(msg.sender) {
        dxtToken = IERC20(_dxtTokenAddress);

        // Inicializar las dos propuestas base del proyecto
        _addProposal(unicode"🤖 NUEVO NFT: ROBOT");
        _addProposal(unicode"💰 SUBIR PAGO MINERO");
    }

    // Función interna para añadir propuestas
    function _addProposal(string memory _description) internal {
        uint256 proposalId = proposals.length;
        proposals.push(Proposal({
            id: proposalId,
            description: _description,
            voteCount: 0,
            votesCountRaw: 0
        }));
        emit ProposalAdded(proposalId, _description);
    }

    // Permitir al dueño crear nuevas propuestas
    function createProposal(string calldata _description) external onlyOwner {
        _addProposal(_description);
    }

    // Votar en una propuesta utilizando el saldo de DXT como poder de voto
    function vote(uint256 _proposalId) external {
        require(_proposalId < proposals.length, "Propuesta no existe");
        require(!hasVoted[_proposalId][msg.sender], "Ya has votado en esta propuesta");

        uint256 weight = dxtToken.balanceOf(msg.sender);
        require(weight > 0, "Debes tener DXT tokens para votar en la gobernanza");

        hasVoted[_proposalId][msg.sender] = true;
        
        Proposal storage prop = proposals[_proposalId];
        prop.voteCount += weight;
        prop.votesCountRaw += 1;

        emit Voted(msg.sender, _proposalId, weight);
    }

    // Obtener el total de propuestas
    function getProposalsCount() external view returns (uint256) {
        return proposals.length;
    }

    // Obtener los detalles de una propuesta
    function getProposal(uint256 _proposalId) external view returns (
        uint256 id,
        string memory description,
        uint256 voteCount,
        uint256 votesCountRaw
    ) {
        require(_proposalId < proposals.length, "Propuesta no existe");
        Proposal memory prop = proposals[_proposalId];
        return (prop.id, prop.description, prop.voteCount, prop.votesCountRaw);
    }
}
