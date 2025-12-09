# Face Recognition Image Finder 🔍

AI-powered tool to automatically find all photos of yourself in large image archives using state-of-the-art face recognition.

## ✨ Features

- 🎯 **High Accuracy**: Uses Facenet512 (99.65% accuracy) with RetinaFace detector
- 📦 **Efficient**: Processes images directly from ZIP without extraction
- 🖼️ **Multiple References**: Use a folder of your photos for better matching
- ⚡ **Real-time Progress**: See matches and progress as it scans
- 🎨 **Format Support**: JPG, PNG, BMP, GIF, TIFF, WEBP
- 🔒 **Privacy**: Runs completely offline on your machine

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.tools.sap/I762933/imag_finder.git
cd imag_finder

# Create virtual environment and install dependencies
python3 -m venv venv
source venv/bin/activate
pip install deepface pillow tf-keras opencv-python
```

### Usage

**With a single reference photo:**
```bash
python find_my_images.py photos.zip my_face.jpg output_folder
```

**With multiple reference photos (recommended):**
```bash
python find_my_images.py photos.zip ./reference_photos/ output_folder
```

**With custom threshold:**
```bash
python find_my_images.py photos.zip ./reference_photos/ output_folder 0.30
```

### Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `zip_file` | Path to ZIP archive containing images | Yes |
| `reference_path` | Path to a reference image OR folder with multiple images | Yes |
| `output_folder` | Folder where matching images will be saved | Yes |
| `threshold` | Face matching threshold (0.0-1.0, default: 0.35) | No |

### Threshold Guide

- **0.25** - Very strict (fewer false positives, may miss some matches)
- **0.35** - Recommended (good balance)
- **0.45** - Lenient (more matches, some false positives)

## 📖 How It Works

1. **Load References**: Extracts face embeddings from your reference photo(s)
2. **Scan Archive**: Reads images directly from ZIP file
3. **Detect Faces**: Uses RetinaFace to find all faces in each image
4. **Compare**: Calculates cosine distance between face embeddings
5. **Match & Save**: Copies matching images to output folder

### Technology Stack

- **Face Detection**: RetinaFace (SOTA detector)
- **Face Recognition**: Facenet512 (512-dimensional embeddings)
- **Distance Metric**: Cosine distance
- **Framework**: DeepFace

## 💡 Tips for Best Results

### Reference Photos
- ✅ Use 3-10 clear photos of yourself
- ✅ Include different angles and expressions
- ✅ Well-lit, front-facing shots work best
- ❌ Avoid sunglasses, masks, or heavy shadows

### Troubleshooting

**Too many false positives?**
- Lower the threshold: `python find_my_images.py ... 0.25`
- Use more reference photos
- Ensure reference photos are clear

**Missing some photos?**
- Increase the threshold: `python find_my_images.py ... 0.40`
- Add more varied reference photos

**First run is slow?**
- Normal! Downloads AI models (~150MB) on first run
- Subsequent runs are much faster

## 🎯 Performance

- **Speed**: ~2-4 seconds per image (depending on hardware)
- **Accuracy**: 99%+ with good reference photos
- **Memory**: ~2GB RAM recommended
- **Storage**: Models require ~150MB

## 📋 Example Session

```bash
$ python find_my_images.py vacation.zip ./my_photos/ matches/

Loading 8 reference image(s)...
✓ Loaded: photo1.jpg
✓ Loaded: photo2.jpg
✓ Loaded: photo3.jpg
...

✓ Successfully loaded 8 reference image(s) with faces

Scanning zip file: vacation.zip
Output folder: matches/
Face matching threshold: 0.35
------------------------------------------------------------
Found 1688 image files to scan

[1/1688] Scanned... (0 matches so far)
[25/1688] ✓ MATCH (dist: 0.287): IMG_1234.jpg
[50/1688] ✓ MATCH (dist: 0.312): IMG_1567.jpg
...

============================================================
Scan complete!
Images scanned: 1688
Matches found: 143
Output folder: matches/
============================================================
```

## 🔧 Advanced Usage

### Custom Model Settings

Edit the script to change models:
```python
# In compare_faces() and verify_reference_faces()
model_name='Facenet512'  # Options: VGG-Face, Facenet, Facenet512, OpenFace, DeepFace, ArcFace
detector_backend='retinaface'  # Options: opencv, retinaface, mtcnn, ssd
```

## 🐛 Known Limitations

- **Obscured faces**: Cannot detect faces hidden by helmets, masks, or heavy shadows
- **Profile shots**: Works best with frontal or near-frontal faces
- **Quality**: Very blurry or low-resolution images may not be detected

## 📄 License

MIT License - See LICENSE file for details

## 🤝 Contributing

Contributions welcome! Please open an issue or PR.

## 📧 Contact

Issues: https://github.tools.sap/I762933/imag_finder/issues
