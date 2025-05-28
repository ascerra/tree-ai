import os
import sys
import argparse
from transformers import AutoTokenizer, AutoModelForCausalLM, GenerationConfig
import torch

# -----------------------
# Argument parsing
# -----------------------
parser = argparse.ArgumentParser()
parser.add_argument("--verbose", action="store_true", help="Enable debug logging")
args = parser.parse_args()
VERBOSE = args.verbose

# -----------------------
# Environment setup
# -----------------------
script_dir = os.path.dirname(os.path.realpath(__file__))
os.environ["HF_HOME"] = os.path.join(script_dir, "..", ".hf-cache")
os.environ["TRANSFORMERS_CACHE"] = os.path.join(script_dir, "..", ".hf-cache")
os.makedirs(os.environ["TRANSFORMERS_CACHE"], exist_ok=True)

# -----------------------
# Load model and tokenizer
# -----------------------
model_name = "ibm-granite/granite-3.1-8b-instruct"
tokenizer = AutoTokenizer.from_pretrained(model_name)
model = AutoModelForCausalLM.from_pretrained(
    model_name,
    torch_dtype=torch.bfloat16,
    device_map="auto"
)
model.eval()

generation_config = GenerationConfig(
    max_new_tokens=100,
    temperature=0.7,
    do_sample=True,
    num_return_sequences=1,
    pad_token_id=tokenizer.eos_token_id,
    eos_token_id=tokenizer.eos_token_id,
)

print("[READY]", flush=True)

# -----------------------
# Main input loop
# -----------------------
while True:
    lines = []
    for line in sys.stdin:
        if line.strip() == "<<END>>":
            break
        lines.append(line)

    full_prompt = "".join(lines).strip()
    if not full_prompt:
        print("[EMPTY]", flush=True)
        continue

    if VERBOSE:
        print("\n[DEBUG] Received prompt:\n" + full_prompt + "\n", file=sys.stderr, flush=True)

    inputs = tokenizer(full_prompt, return_tensors="pt", padding=False)
    input_ids = inputs["input_ids"].to(model.device)
    attention_mask = inputs["attention_mask"].to(model.device)

    if VERBOSE:
        print(f"[DEBUG] Prompt token count: {input_ids.shape[1]}", file=sys.stderr, flush=True)

    outputs = model.generate(
        input_ids=input_ids,
        attention_mask=attention_mask,
        generation_config=generation_config
    )

    generated_text = tokenizer.decode(
        outputs[0][input_ids.shape[1]:],
        skip_special_tokens=True
    )
    print(generated_text.strip(), flush=True)
