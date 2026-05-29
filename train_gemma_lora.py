import os
import torch
from datasets import load_dataset
from transformers import AutoTokenizer, AutoModelForCausalLM, BitsAndBytesConfig, TrainingArguments
from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
from trl import SFTTrainer

# Model configurations
MODEL_ID = "google/gemma-2b"
DATASET_PATH = "terminal_train.jsonl"
OUTPUT_DIR = "./gemma-lora-output"

def train():
    print("🚀 Cargando tokenizador y modelo Gemma base...")
    
    # 4-bit quantization configuration for QLoRA
    bnb_config = BitsAndBytesConfig(
        load_in_4bit=True,
        bnb_4bit_quant_type="nf4",
        bnb_4bit_compute_dtype=torch.float16,
        bnb_4bit_use_double_quant=True
    )
    
    tokenizer = AutoTokenizer.from_pretrained(MODEL_ID)
    tokenizer.pad_token = tokenizer.eos_token
    
    # Load causal language model
    model = AutoModelForCausalLM.from_pretrained(
        MODEL_ID,
        quantization_config=bnb_config,
        device_map="auto"
    )
    
    # Prepare model for training in 8bit/4bit
    model = prepare_model_for_kbit_training(model)
    
    # Configure LoRA settings
    lora_config = LoraConfig(
        r=8,
        lora_alpha=16,
        target_modules=["q_proj", "o_proj", "k_proj", "v_proj", "gate_proj", "up_proj", "down_proj"],
        lora_dropout=0.05,
        bias="none",
        task_type="CAUSAL_LM"
    )
    model = get_peft_model(model, lora_config)
    
    print(f"📦 Cargando dataset de entrenamiento desde: {DATASET_PATH}...")
    dataset = load_dataset("json", data_files=DATASET_PATH, split="train")
    
    # Simple formatting prompt function
    def formatting_prompts_func(example):
        output_texts = []
        for i in range(len(example['prompt'])):
            text = f"### Instruction:\n{example['prompt'][i]}\n\n### Response:\n{example['command'][i]}"
            output_texts.append(text)
        return output_texts

    # Set up training arguments
    training_args = TrainingArguments(
        output_dir=OUTPUT_DIR,
        per_device_train_batch_size=2,
        gradient_accumulation_steps=4,
        warmup_steps=10,
        max_steps=100,
        learning_rate=2e-4,
        fp16=True,
        logging_steps=1,
        save_strategy="steps",
        save_steps=50,
        optim="paged_adamw_8bit"
    )
    
    # SFT Trainer initialization
    trainer = SFTTrainer(
        model=model,
        train_dataset=dataset,
        peft_config=lora_config,
        max_seq_length=512,
        tokenizer=tokenizer,
        formatting_func=formatting_prompts_func,
        args=training_args,
    )
    
    print("⛏️ Iniciando ajuste fino Gemma con QLoRA en la GPU...")
    trainer.train()
    
    # Save the custom LoRA adapter weights
    print(f"💾 Guardando adaptador LoRA en: {OUTPUT_DIR}...")
    trainer.model.save_pretrained(OUTPUT_DIR)
    print("🎉 ¡Ajuste fino Gemma completado con éxito!")

if __name__ == "__main__":
    train()
