import os

# Set local cache paths before importing transformers
script_dir = os.path.dirname(os.path.realpath(__file__))
os.environ["HF_HOME"] = os.path.join(script_dir, "..", ".hf-cache")
os.environ["TRANSFORMERS_CACHE"] = os.path.join(script_dir, "..", ".hf-cache")
os.makedirs(os.environ["TRANSFORMERS_CACHE"], exist_ok=True)
import sys
from transformers import AutoModelForCausalLM, AutoTokenizer
import torch

def main():
    if len(sys.argv) < 2:
        print("Usage: python granite_infer.py --prompt '<prompt>'", file=sys.stderr)
        sys.exit(1)

    # Use local Hugging Face cache directory
    script_dir = os.path.dirname(os.path.realpath(__file__))
    os.environ["HF_HOME"] = os.path.join(script_dir, "..", ".hf-cache")

    prompt = " ".join(sys.argv[2:]) if sys.argv[1] == "--prompt" else sys.argv[1]

    model_name = "ibm-granite/granite-3.1-8b-instruct"
    tokenizer = AutoTokenizer.from_pretrained(model_name)
    model = AutoModelForCausalLM.from_pretrained(model_name)

    inputs = tokenizer(prompt, return_tensors="pt")
    outputs = model.generate(**inputs, max_new_tokens=100)
    response = tokenizer.decode(outputs[0], skip_special_tokens=True)
    print(response)

if __name__ == "__main__":
    main()