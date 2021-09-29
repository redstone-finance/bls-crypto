/*
 * Copyright 2020 ConsenSys Software Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
pragma solidity >=0.7.1;

import "./BlsSignatureVerification.sol";


contract BlsSignatureTest is BlsSignatureVerification {
    bool public verified;

    function verifySignature(
        bytes calldata _publicKey,  // an E2 point
        bytes calldata _message,
        bytes calldata _signature   // an E1 point
    ) external {
        E2Point memory pub = decodePublicKey(_publicKey);
        E1Point memory sig = decodeSignature(_signature);
        verified = verify(pub, _message, sig);
    }

    function verifySignaturePoint(
        bytes calldata _publicKey,  // an E2 point
        bytes calldata _message,    // an E1 point
        bytes calldata _signature   // an E1 point
    ) external {
        E2Point memory pub = decodePublicKey(_publicKey);
        E1Point memory sig = decodeSignature(_signature);
        verified = verifyForPoint(pub, decodeSignature(_message), sig);
    }

    function verifyAggregatedSignature(
        bytes calldata _aggregatedPublicKey,  // an E2 point
        bytes calldata _partPublicKey,        // an E2 point
        bytes calldata _message,
        bytes calldata _partSignature,        // an E1 point
        uint _signersBitmask
    ) external {
        E2Point memory aPub = decodePublicKey(_aggregatedPublicKey);
        E2Point memory pPub = decodePublicKey(_partPublicKey);
        E1Point memory pSig = decodeSignature(_partSignature);
        verified = verifyAggregated(aPub, pPub, _message, pSig, _signersBitmask);
    }

    function verifyAggregatedHash(
        bytes calldata _p,
        uint index
    ) external view returns (bytes memory) {
        E2Point memory pub = decodePublicKey(_p);
        bytes memory message = abi.encodePacked(pub.x, pub.y, index);
        E1Point memory h = hashToCurveE1(message);
        return abi.encodePacked(h.x, h.y);
    }

    function addOnCurveE1(
        bytes calldata _p1,
        bytes calldata _p2
    ) external view returns (bytes memory) {
        E1Point memory res = addCurveE1(decodeSignature(_p1), decodeSignature(_p2));
        return abi.encode(res.x, res.y);
    }

    function decodePublicKey(bytes memory _pubKey) private pure returns (E2Point memory pubKey) {
        uint256[] memory output = new uint256[](4);
        for (uint256 i=32; i<=output.length*32; i+=32) {
            assembly { mstore(add(output, i), mload(add(_pubKey, i))) }
        }

        pubKey.x[0] = output[0];
        pubKey.x[1] = output[1];
        pubKey.y[0] = output[2];
        pubKey.y[1] = output[3];
    }

    function decodeSignature(bytes memory _sig) private pure returns (E1Point memory signature) {
        uint256[] memory output = new uint256[](2);
        for (uint256 i=32; i<=output.length*32; i+=32) {
            assembly { mstore(add(output, i), mload(add(_sig, i))) }
        }

        signature = E1Point(0, 0);
        signature.x = output[0];
        signature.y = output[1];
    }
}
