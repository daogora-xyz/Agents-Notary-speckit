#import "@preview/messeji:0.3.0": messeji

#set page(width: 21cm, height: auto, margin: 2cm)
#set text(font: "Helvetica Neue", size: 11pt)

#align(center)[
  #text(size: 18pt, weight: "bold")[
    Agents Notary Platform
  ]
  #v(0.5em)
  #text(size: 14pt)[
    Document Certification Flow
  ]
  #v(0.3em)
  #text(size: 10pt, fill: gray)[
    Example: Certifying "Nakamoto Satoshi - CV/Resume \[English\].pdf"
  ]
]

#v(2em)

#let certification-chat = (
  // Initial request
  (
    date: "2025-10-30T14:20:15",
    msg: "Hi! I need to permanently certify my resume on the blockchain for a job application.\n\nFile: Nakamoto Satoshi - CV/Resume [English].pdf\nSHA-256: 0x3f2a8d9c1b4e7f6a2d8c9b3e4f5a6d7c8b9a0e1f2d3c4b5a6d7e8f9a0b1c2d3e",
    from_me: true,
  ),

  (
    msg: "I can help you certify that resume on the Circular Protocol blockchain! 🔐\n\nThis creates an immutable, timestamped proof that anyone can verify.\n\nCost: \\$0.50 USDC on Base network\nBlockchain fee: 4 CIRX (included)\n\nReady to proceed?",
    from_me: false,
  ),

  (
    msg: "Yes, let's do it. How does payment work?",
    from_me: true,
  ),

  // Payment phase
  (
    msg: "[Calling: create_payment_requirement]\n\n💳 Payment Request Generated\n\nAmount: 0.50 USDC\nNetwork: Base (mainnet)\nTo: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb\n\nPlease sign this payment authorization with your wallet.",
    from_me: false,
  ),

  (
    msg: "[Wallet Connected]\n\nSigned! Here's my payment authorization:\n0x8a2f3d4c5b6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a",
    from_me: true,
  ),

  (
    msg: "[Calling: verify_payment]\n\nVerifying signature...\n✅ Payment signature verified!\n\n[Calling: settle_payment]\n\nSubmitting to x402 facilitator...",
    from_me: false,
  ),

  (
    msg: "✅ Payment settled on Base blockchain!\n\nTransaction confirmed in 8 seconds.\n\nNow certifying your resume on Circular Protocol...",
    from_me: false,
  ),

  // Certification phase
  (
    date: "2025-10-30T14:20:45",
    msg: "[Calling: get_wallet_nonce]\n\nRetrieving nonce: 42\n\n[Calling: certify_data]\n\nBuilding C_TYPE_CERTIFICATE transaction...\nCalculating TX ID: SHA-256(Blockchain+From+To+Payload+Nonce+Timestamp)...\nSigning with Secp256k1...\nSubmitting to Circular Protocol...",
    from_me: false,
  ),

  (
    msg: "📤 Certification transaction submitted!\n\nTransaction ID: tx_8f3d2a1c9e7b5f4d6a8c2e1f9b3d7a5c4e6f8a9b0c1d2e3f4a5b6c7d8e9f0a1b\n\nMonitoring status...",
    from_me: false,
  ),

  (
    msg: "How long will this take?",
    from_me: true,
  ),

  // Status polling phase
  (
    msg: "[Calling: get_transaction_status]\n\nPolling with exponential backoff...\n\n⏳ Status: Pending (in mempool)\nWait 2 seconds...",
    from_me: false,
  ),

  (
    msg: "🔄 Status: Verified (included in block)\nWait 3 seconds...",
    from_me: false,
  ),

  (
    msg: "✅ Status: Executed (finalized on blockchain!)\n\nTotal certification time: 18 seconds",
    from_me: false,
    reaction: "🎉",
  ),

  // Proof generation phase
  (
    date: "2025-10-30T14:21:03",
    msg: "[Calling: get_certification_proof]\n\nGenerating verifiable proof...",
    from_me: false,
  ),

  (
    msg: "🎉 Certification Complete!\n\n📜 Proof Details:\n━━━━━━━━━━━━━━━━━━━━━━━━━━\nDocument: Nakamoto Satoshi - CV/Resume [English].pdf\nSHA-256: 0x3f2a8d9c1b4e7f6a...\n\nTransaction ID:\ntx_8f3d2a1c9e7b5f4d...\n\nBlock ID:\nblk_9d4e2f1a8c7b6d5e...\n\nTimestamp:\n2025-10-30 14:21:03 UTC\n\nNetwork:\nCircular Protocol Mainnet\n━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n🔗 View on Explorer:\nhttps://circularlabs.io/Explorer?network=mainnet&tx=tx_8f3d2a1c...",
    from_me: false,
  ),

  (
    msg: "Perfect! Can I share this proof with the employer?",
    from_me: true,
    reaction: "😊",
  ),

  (
    msg: "Absolutely! You can share:\n\n1. The explorer URL (fully verifiable by anyone)\n2. The transaction ID (they can look it up independently)\n3. The full proof JSON (contains all metadata)\n\nThe employer can verify your resume certification without trusting anyone - the blockchain is the source of truth! 🔐\n\nYour resume is now permanently timestamped and cannot be backdated or altered.",
    from_me: false,
  ),

  (
    msg: "This is amazing! The employer can verify I created this resume before today. Thank you!",
    from_me: true,
    reaction: "🙏",
  ),

  (
    msg: "You're welcome! Your certification is now part of the immutable blockchain record. Good luck with your job application! 🚀",
    from_me: false,
    reaction: "❤️",
  ),
)

#messeji(
  chat-data: certification-chat,
  timestamp-format: "[hour]:[minute]:[second]",
  date-changed-format: "[year]-[month]-[day] [hour]:[minute]",
)

#v(2em)
#line(length: 100%, stroke: 0.5pt + gray)
#v(1em)

#text(size: 9pt, fill: gray)[
  *System Components Used:*
  - x402-mcp-server: Payment processing (USDC on Base)
  - circular-mcp-server: Blockchain certification (Circular Protocol)
  - Total time: ~48 seconds (8s payment + 18s certification + 22s processing)
  - Cost: \$0.50 USDC + 4 CIRX blockchain fee
]
