--CONTEXT: LLM/AGENT HAS NO WALLET/PRIVATE KEYS AND ISN'T AN MCP HOST (CAN'T USE TOOLS EXPOSED BY MCP SERVERS NATIVELY--

User Prompt: "Hey, could you help me certify 'ashley-barr_offerletter-contract--unsigned.pdf' using certify.ar4s.com? The file should be attached"
LLM/Agent (Reply): "Sure, no problem!"
  /// The LLM/Agent should query the endpoint and the back and forth should represent the discovery process for the calling llm/agent, either via helpful/guiding error messages or another means. The first part of the workflow will have the calling llm discover that the certify endpoint isn't available, as its a paid resource, and that the first requirement needs to be fulfilled, namely, it needs to upload the document so it can recieve a quote, so it can move on to the next step: payment to access the payment gated resource /certify.
  LLM/Agent [Action] -> Runs <Action> against endpoint via <Protocol>
  LLM/Agent [Receives] <- <Response>
    LLM/Aagent <-> Service {Negotiate to allow discovery of services and exposure of gated MCP tools}
  LLM/Agent {Action} -> Uploads document.
  LLM/Agent [Receives] <- <Response with quote, x402 payment instructions for agent>
LLM/Agent (Reply): "To certify your <size in KB/MB> <filename>.<ext> document it will cost $00.XX USDC ($00.XX Processing Fee + 4 CIRX [@<exchange rate>]), Proceed?
User Prompt: "Yes, please"
LLM/Agent (Reply): "Great, let me try."
  LLM/Agent {Action} -> Tries to send a tx as per instructions given from the service during the previous step, but isn't able to. 
  LLM/Agent [Receives} <- ...
