#!/usr/bin/env python3
"""
Face Recognition Script - Find Your Images in a Zip Archive

This script scans all images in a zip file and identifies images containing your face
by comparing them against a reference image of yourself.

Requirements:
    pip install deepface pillow tf-keras opencv-python

Usage:
    python find_my_images.py <zip_file_path> <reference_path> <output_folder>

Example (single image):
    python find_my_images.py photos.zip my_face.jpg output_folder
    
Example (folder of images):
    python find_my_images.py photos.zip ./my_photos_folder output_folder
"""

import sys
import os
import zipfile
from pathlib import Path
from PIL import Image
import io
from typing import List
from deepface import DeepFace
import cv2
import numpy as np


def is_image_file(filename: str) -> bool:
    """Check if a file is an image based on its extension."""
    image_extensions = {'.jpg', '.jpeg', '.png', '.bmp', '.gif', '.tiff', '.webp'}
    return Path(filename).suffix.lower() in image_extensions


def load_reference_images(reference_path: str) -> List[str]:
    """
    Load reference images from a file or folder.
    
    Args:
        reference_path: Path to a single image file or a folder containing multiple images
    
    Returns:
        List of paths to valid reference images
    """
    reference_images = []
    
    if not os.path.exists(reference_path):
        raise FileNotFoundError(f"Reference path not found: {reference_path}")
    
    # Check if it's a file or directory
    if os.path.isfile(reference_path):
        if is_image_file(reference_path):
            reference_images.append(reference_path)
        else:
            raise ValueError(f"File is not a supported image format: {reference_path}")
    
    elif os.path.isdir(reference_path):
        # Load all image files from the directory
        for filename in sorted(os.listdir(reference_path)):
            file_path = os.path.join(reference_path, filename)
            # Skip hidden files and non-files
            if filename.startswith('.') or not os.path.isfile(file_path):
                continue
            if is_image_file(filename):
                reference_images.append(file_path)
    
    if not reference_images:
        raise ValueError(f"No valid image files found in: {reference_path}")
    
    return reference_images


def verify_reference_faces(reference_images: List[str]) -> tuple[List[str], List[np.ndarray]]:
    """
    Verify that the reference images contain detectable faces and extract embeddings.
    
    Returns:
        Tuple of (valid_reference_paths, reference_embeddings)
    """
    print(f"\nLoading {len(reference_images)} reference image(s)...")
    
    valid_references = []
    reference_embeddings = []
    
    for img_path in reference_images:
        try:
            # Extract face embedding from reference image
            embedding_objs = DeepFace.represent(
                img_path=img_path,
                model_name='Facenet512',
                detector_backend='retinaface',
                enforce_detection=True
            )
            
            if len(embedding_objs) > 0:
                # Use the first (most prominent) face
                embedding = embedding_objs[0]['embedding']
                valid_references.append(img_path)
                reference_embeddings.append(np.array(embedding))
                print(f"✓ Loaded: {os.path.basename(img_path)}")
            else:
                print(f"✗ Skipped (no face): {os.path.basename(img_path)}")
                
        except Exception as e:
            print(f"✗ Skipped (error): {os.path.basename(img_path)}")
            continue
    
    if not valid_references:
        raise ValueError("No faces found in any reference images. Please provide clear images with your face.")
    
    print(f"\n✓ Successfully loaded {len(valid_references)} reference image(s) with faces")
    return valid_references, reference_embeddings


def compare_faces(img1_array: np.ndarray, reference_embeddings: List[np.ndarray], threshold: float = 0.35) -> tuple[bool, float]:
    """
    Compare a face in an image with multiple reference embeddings.
    Returns True if the face matches ANY of the reference images.
    
    Args:
        img1_array: Numpy array of the image to check
        reference_embeddings: List of reference face embeddings
        threshold: Cosine distance threshold (lower is more strict, 0-1)
    
    Returns:
        Tuple of (is_match, best_distance)
    """
    try:
        # Extract face embeddings from the test image
        embedding_objs = DeepFace.represent(
            img_path=img1_array,
            model_name='Facenet512',
            detector_backend='retinaface',
            enforce_detection=False
        )
        
        if not embedding_objs:
            return False, 1.0
        
        best_distance = 1.0
        
        # Check all faces in the image
        for embedding_obj in embedding_objs:
            test_embedding = np.array(embedding_obj['embedding'])
            
            # Compare with each reference embedding
            for ref_embedding in reference_embeddings:
                # Calculate cosine distance
                distance = np.dot(test_embedding, ref_embedding) / (
                    np.linalg.norm(test_embedding) * np.linalg.norm(ref_embedding)
                )
                # Convert similarity to distance (1 - similarity)
                distance = 1 - distance
                
                if distance < best_distance:
                    best_distance = distance
                
                # If we find a good match, return early
                if distance < threshold:
                    return True, distance
        
        return False, best_distance
        
    except Exception as e:
        return False, 1.0


def scan_zip_for_faces(zip_path: str, reference_embeddings: List[np.ndarray], output_folder: str, threshold: float = 0.35):
    """
    Scan all images in a zip file and extract those containing the reference face.
    
    Args:
        zip_path: Path to the zip file
        reference_embeddings: List of reference face embeddings
        output_folder: Folder to save matching images
        threshold: Face matching threshold (lower is more strict, default 0.35)
    """
    if not os.path.exists(zip_path):
        raise FileNotFoundError(f"Zip file not found: {zip_path}")
    
    # Create output folder if it doesn't exist
    os.makedirs(output_folder, exist_ok=True)
    
    print(f"\nScanning zip file: {zip_path}")
    print(f"Output folder: {output_folder}")
    print(f"Face matching threshold: {threshold}")
    print("-" * 60)
    
    matches_found = 0
    images_scanned = 0
    total_files = 0
    
    with zipfile.ZipFile(zip_path, 'r') as zip_ref:
        # Get list of all image files in the zip
        all_files = zip_ref.namelist()
        image_files = [f for f in all_files if is_image_file(f) and not f.startswith('__MACOSX')]
        total_files = len(image_files)
        
        print(f"Found {total_files} image files to scan\n")
        
        for idx, file_path in enumerate(image_files, 1):
            try:
                # Read image from zip
                with zip_ref.open(file_path) as image_file:
                    image_data = image_file.read()
                    
                # Load image using PIL and convert to format for processing
                image = Image.open(io.BytesIO(image_data))
                
                # Convert to RGB if necessary
                if image.mode != 'RGB':
                    image = image.convert('RGB')
                
                # Convert PIL image to numpy array
                image_array = np.array(image)
                
                # Convert RGB to BGR for OpenCV
                image_array = cv2.cvtColor(image_array, cv2.COLOR_RGB2BGR)
                
                images_scanned += 1
                
                # Check if face matches any of the reference faces
                match_found, distance = compare_faces(image_array, reference_embeddings, threshold)
                
                if match_found:
                    # Save the matching image
                    output_path = os.path.join(output_folder, os.path.basename(file_path))
                    
                    # Handle duplicate filenames
                    if os.path.exists(output_path):
                        base, ext = os.path.splitext(os.path.basename(file_path))
                        counter = 1
                        while os.path.exists(output_path):
                            output_path = os.path.join(output_folder, f"{base}_{counter}{ext}")
                            counter += 1
                    
                    with open(output_path, 'wb') as out_file:
                        out_file.write(image_data)
                    
                    matches_found += 1
                    print(f"[{idx}/{total_files}] ✓ MATCH (dist: {distance:.3f}): {file_path}")
                else:
                    # Print progress every 20 images
                    if idx % 20 == 0:
                        print(f"[{idx}/{total_files}] Scanned... ({matches_found} matches so far)")
                
            except Exception as e:
                print(f"[{idx}/{total_files}] ✗ Error processing {file_path}: {str(e)}")
                continue
    
    print("\n" + "=" * 60)
    print("Scan complete!")
    print(f"Images scanned: {images_scanned}")
    print(f"Matches found: {matches_found}")
    print(f"Output folder: {output_folder}")
    print("=" * 60)


def main():
    """Main function to parse arguments and run the face recognition scan."""
    if len(sys.argv) < 4:
        print(__doc__)
        print("\nError: Missing required arguments")
        print("\nUsage:")
        print(f"  python {sys.argv[0]} <zip_file> <reference_path> <output_folder> [threshold]")
        print("\nArguments:")
        print("  zip_file        : Path to the zip file containing images")
        print("  reference_path  : Path to a reference image OR folder containing multiple images of yourself")
        print("  output_folder   : Folder where matching images will be saved")
        print("  threshold       : (Optional) Face matching threshold, 0.0-1.0 (default: 0.35)")
        print("                    Lower values are more strict (0.35 recommended, 0.25 very strict)")
        sys.exit(1)
    
    zip_path = sys.argv[1]
    reference_path = sys.argv[2]
    output_folder = sys.argv[3]
    threshold = float(sys.argv[4]) if len(sys.argv) > 4 else 0.35
    
    try:
        # Load reference images (file or folder)
        reference_images = load_reference_images(reference_path)
        
        # Verify reference faces and extract embeddings
        valid_references, reference_embeddings = verify_reference_faces(reference_images)
        
        # Scan zip file for matching faces
        scan_zip_for_faces(zip_path, reference_embeddings, output_folder, threshold)
        
    except Exception as e:
        print(f"\n✗ Error: {str(e)}")
        sys.exit(1)


if __name__ == "__main__":
    main()
