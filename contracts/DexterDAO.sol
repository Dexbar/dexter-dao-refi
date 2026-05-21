// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract DexterDAO is ERC20 {
    address public owner;
    constructor(uint256 initialSupply) ERC20("Dexter DAO Token", "DXT") {
        owner = msg.sender;
        _mint(owner, initialSupply);
    }
    // Mint function for testing (only owner)
    function mint(address to, uint256 amount) external {
        require(msg.sender == owner, "Only owner can mint");
        _mint(to, amount);
    }
}
