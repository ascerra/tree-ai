import os
import sys
from transformers import AutoTokenizer, AutoModelForCausalLM, GenerationConfig
import torch

# Set local cache paths before importing transformers
script_dir = os.path.dirname(os.path.realpath(__file__))
os.environ["HF_HOME"] = os.path.join(script_dir, "..", ".hf-cache")
os.environ["TRANSFORMERS_CACHE"] = os.path.join(script_dir, "..", ".hf-cache")
os.makedirs(os.environ["TRANSFORMERS_CACHE"], exist_ok=True)

model_name = "ibm-granite/granite-3.1-8b-instruct"

print(f"Loading tokenizer for {model_name}...")
tokenizer = AutoTokenizer.from_pretrained(model_name)
print("Tokenizer loaded.")

print(f"Loading model {model_name}...")
model = AutoModelForCausalLM.from_pretrained(
    model_name,
    torch_dtype=torch.bfloat16, # Or torch.float16 if bfloat16 is not supported
    device_map="auto"
)
print("Model loaded.")

model.eval()

# Define a sample prompt similar to what Go would send
sample_file_content = """
// go.mod
module tree-ai

go 1.20

require github.com/spf13/cobra v1.6.1

require (
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
"""

item_type = "file"
target = "go.mod"
instruction = f"In 1 sentence, explain the purpose of this {item_type} **as it relates to the whole project**. Respond only with the explanation. Avoid repeating the file name or type."

# This is the full prompt string that the Go app would send
full_prompt_from_go = f"""Given the following {item_type} named "{target}" with its contents:
{sample_file_content}

{instruction}"""

# Add the explicit "Response:" cue for the model
prompt_for_granite = full_prompt_from_go + "\n\nResponse:"

print("\n--- PROMPT BEING SENT TO MODEL ---")
print(prompt_for_granite)
print("----------------------------------\n")


messages = [
    {"role": "user", "content": prompt_for_granite}
]

# Apply chat template to get the tokenized string
formatted_prompt = tokenizer.apply_chat_template(messages, tokenize=False, add_generation_prompt=True)

# Tokenize the formatted prompt directly, ensuring return_tensors="pt"
inputs = tokenizer(formatted_prompt, return_tensors="pt")

# Move inputs to model's device
input_ids = inputs['input_ids'].to(model.device)
attention_mask = inputs['attention_mask'].to(model.device)

# Define generation configuration
generation_config = GenerationConfig(
    max_new_tokens=100,
    temperature=0.7,
    do_sample=True,
    num_return_sequences=1,
    pad_token_id=tokenizer.eos_token_id,
    eos_token_id=tokenizer.eos_token_id,
)

print("Generating response...")
outputs = model.generate(
    input_ids=input_ids,
    attention_mask=attention_mask,
    generation_config=generation_config
)
print("Response generated.")

# Decode the response.
generated_sequence = outputs[0]

# Calculate the start index of the generated text
start_index = len(input_ids[0])

# Decode only the generated part
generated_text = tokenizer.decode(generated_sequence[start_index:], skip_special_tokens=True)

# Post-process: remove the "Response:" or similar markers from the output if the model echoed them
generated_text = generated_text.replace("Response:", "").strip()

print("\n--- GENERATED RESPONSE ---")
print(generated_text)
print("--------------------------\n")