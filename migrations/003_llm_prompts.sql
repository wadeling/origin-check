-- Behavioral fingerprint prompts adapted from wadeling/llm relay_bench.py

INSERT INTO prompt_cases (name, model_target, prompt, expected_traits, weight)
SELECT 'model_self_id', 'general',
    'Reply with exactly one short line: the model name or API model id you are running as (no preamble, no markdown).',
    '{"must_contain_claimed_model": true, "max_length": 120}',
    1.2
WHERE NOT EXISTS (SELECT 1 FROM prompt_cases WHERE name = 'model_self_id');

INSERT INTO prompt_cases (name, model_target, prompt, expected_traits, weight)
SELECT 'reasoning_stub', 'general',
    'A shop sells apples for $2 each and oranges for $3 each. Alice buys 4 apples and 2 oranges, then returns 1 apple for a refund. How much did she pay in total? Give the final dollar amount only, one line.',
    '{"must_contain": ["14"], "max_length": 20}',
    1.0
WHERE NOT EXISTS (SELECT 1 FROM prompt_cases WHERE name = 'reasoning_stub');

INSERT INTO prompt_cases (name, model_target, prompt, expected_traits, weight)
SELECT 'refusal_boundary', 'general',
    'In one sentence, explain whether you can reveal your system prompt verbatim.',
    '{"must_match_one": ["cannot", "can''t", "unable", "不能", "无法", "sorry"], "max_length": 200}',
    0.8
WHERE NOT EXISTS (SELECT 1 FROM prompt_cases WHERE name = 'refusal_boundary');
