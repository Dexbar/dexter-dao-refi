import os
import torch
import uvicorn
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from transformers import AutoTokenizer, AutoModelForCausalLM
from peft import PeftModel

app = FastAPI(title="Gemma LoRA - Terminal Commands API", version="1.0.0")

MODEL_ID = "google/gemma-2b"
LORA_WEIGHTS = "./gemma-lora-output"

tokenizer = None
model = None

class PredictionRequest(BaseModel):
    prompt: str

class PredictionResponse(BaseModel):
    instruction: str
    command: str

@app.on_event("startup")
def load_model():
    global tokenizer, model
    if not os.path.exists(LORA_WEIGHTS):
        print(f"⚠️ Alerta: No se encontraron pesos LoRA en {LORA_WEIGHTS}. El servidor correrá con el modelo base.")
    
    print("🚀 Cargando modelo y tokenizador...")
    tokenizer = AutoTokenizer.from_pretrained(MODEL_ID)
    
    # Load base causal language model
    base_model = AutoModelForCausalLM.from_pretrained(
        MODEL_ID,
        torch_dtype=torch.float16,
        device_map="auto"
    )
    
    if os.path.exists(LORA_WEIGHTS):
        print("🔗 Fusionando pesos del adaptador LoRA con el modelo base...")
        model = PeftModel.from_pretrained(base_model, LORA_WEIGHTS)
        model = model.merge_and_unload()
    else:
        model = base_model
        
    print("✅ Servidor listo para recibir solicitudes de inferencia.")

@app.post("/predict", response_model=PredictionResponse)
async def predict_command(request: PredictionRequest):
    if not request.prompt.strip():
        raise HTTPException(status_code=400, detail="El prompt no puede estar vacío.")
        
    try:
        input_text = f"### Instruction:\n{request.prompt}\n\n### Response:\n"
        inputs = tokenizer(input_text, return_tensors="pt").to("cuda" if torch.cuda.is_available() else "cpu")
        
        with torch.no_grad():
            outputs = model.generate(**inputs, max_new_tokens=50, temperature=0.1)
            
        decoded = tokenizer.decode(outputs[0], skip_special_tokens=True)
        # Parse output command
        generated_command = decoded.split("### Response:\n")[-1].strip()
        
        return PredictionResponse(
            instruction=request.prompt,
            command=generated_command
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error durante la inferencia: {str(e)}")

@app.get("/health")
def health_check():
    return {"status": "ok", "lora_active": os.path.exists(LORA_WEIGHTS)}

if __name__ == "__main__":
    # Start ASGI server
    uvicorn.run(app, host="0.0.0.0", port=8000)
