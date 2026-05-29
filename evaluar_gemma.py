import os
import torch
from transformers import AutoTokenizer, AutoModelForCausalLM
from peft import PeftModel

MODEL_ID = "google/gemma-2b"
LORA_WEIGHTS = "./gemma-lora-output"

def evaluar():
    print("🚀 Cargando modelo base Gemma y adaptador LoRA...")
    tokenizer = AutoTokenizer.from_pretrained(MODEL_ID)
    
    # Load base causal language model
    base_model = AutoModelForCausalLM.from_pretrained(
        MODEL_ID,
        torch_dtype=torch.float16,
        device_map="auto"
    )
    
    # Load Fine-Tuned model weights (PEFT Adapter)
    print(f"🔗 Fusionando pesos del adaptador desde {LORA_WEIGHTS}...")
    model = PeftModel.from_pretrained(base_model, LORA_WEIGHTS)
    model = model.merge_and_unload()
    print("✅ Fusión de pesos completada con éxito.")
    
    # Validation test prompts
    test_prompts = [
        "compile the smart contracts",
        "deploy contracts to local network",
        "run the go web server",
        "check dxt token balance of local account"
    ]
    
    print("\n🔍 Ejecutando pruebas de inferencia para validación:")
    print("-------------------------------------------------------")
    for prompt in test_prompts:
        input_text = f"### Instruction:\n{prompt}\n\n### Response:\n"
        inputs = tokenizer(input_text, return_tensors="pt").to("cuda" if torch.cuda.is_available() else "cpu")
        
        with torch.no_grad():
            outputs = model.generate(**inputs, max_new_tokens=50, temperature=0.1)
            
        decoded = tokenizer.decode(outputs[0], skip_special_tokens=True)
        # Extract the assistant command output
        response = decoded.split("### Response:\n")[-1].strip()
        print(f"Instrucción:  '{prompt}'")
        print(f"Comando Gen:  '{response}'")
        print("-------------------------------------------------------")

if __name__ == "__main__":
    if not os.path.exists(LORA_WEIGHTS):
        print(f"❌ No se encontró el directorio de adaptadores: '{LORA_WEIGHTS}'.")
        print("💡 Sugerencia: Ejecuta primero train_gemma_lora.py para entrenar el modelo.")
    else:
        evaluar()
