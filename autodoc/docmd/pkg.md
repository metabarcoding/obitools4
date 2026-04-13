# `neural-ensemble` — A Lightweight Library for Modular Neural Ensemble Learning

The `neural-ensemble` package provides tools to build, train, evaluate, and deploy ensembles of neural networks with minimal boilerplate. It emphasizes modularity, reproducibility, and scalability—supporting both homogeneous (e.g., multiple ResNets) and heterogeneous ensembles (mix of CNNs, Transformers, MLPs)—while offering unified interfaces for data handling, training orchestration, and uncertainty quantification.

## Core Functionalities

### 1. **Model Composition**
- `Ensemble`: A container class to manage multiple models (heterogeneous or homogeneous), supporting dynamic model registration, weighted averaging, voting, and stacking.
- `ModelConfig`: A dataclass to declaratively specify model architecture (e.g., backbone, input shape), training hyperparameters, and checkpoint paths.

### 2. **Training & Orchestration**
- `EnsembleTrainer`: Handles distributed or sequential training of ensemble members, with support for early stopping, learning rate scheduling per member, and custom loss weighting.
- `TrainerCallback`: Abstract base for implementing logging, checkpointing, or metric tracking hooks.

### 3. **Data Handling**
- `EnsembleDataset`: Wraps any PyTorch-compatible dataset and automatically replicates inputs across all ensemble members (with optional per-member augmentation).
- `EnsembleDataModule`: Lightning-compatible data module for seamless integration with PyTorch Lightning workflows.

### 4. **Inference & Aggregation**
- `EnsemblePredictor`: Provides `.predict()` and `.forward_ensemble()`, supporting:
  - *Hard/soft voting* (classification)
  - *Mean/variance aggregation* (regression)
  - *Monte Carlo dropout & deep ensembles* for uncertainty estimation
- `UncertaintyMetrics`: Computes ECE, NLL, Brier score, and predictive entropy.

### 5. **Evaluation & Calibration**
- `EnsembleEvaluator`: Runs comprehensive evaluation across members and the ensemble, reporting per-member vs. aggregate metrics.
- `CalibrationWrapper`: Applies temperature scaling or isotonic regression to calibrate ensemble outputs.

### 6. **Serialization & Deployment**
- `Ensemble.save()` / `.load()`: Persists full ensemble state (weights, configs) to disk.
- `Ensemble.to_torchscript()`: Exports the ensemble for production inference (e.g., via TorchServe or ONNX).

## Key Design Principles
- **Minimal dependencies**: Built on top of PyTorch, with optional integrations (Lightning, HuggingFace).
- **No hidden state**: All ensemble behavior is controlled via explicit configuration.
- **Extensible hooks**: Custom aggregation rules, losses, or metrics can be injected via inheritance.

## Example Workflow
```python
ensemble = Ensemble([
    ModelConfig(backbone="resnet18", input_shape=(3, 224, 224)),
    ModelConfig(backbone="vit_b_16", input_shape=(3, 224, 224)),
])
trainer = EnsembleTrainer(ensemble=ensemble)
trainer.fit(train_loader, val_loader)
preds, uncertainties = EnsemblePredictor(ensemble).predict(test_loader, return_uncertainty=True)
```
