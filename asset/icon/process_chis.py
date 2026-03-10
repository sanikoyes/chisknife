#!/usr/bin/env python3
"""处理 chis.png - 移除右下角水印并添加透明通道"""

from PIL import Image
import numpy as np

# 打开图片
img = Image.open('chis.png')
print(f"原始图片尺寸: {img.size}, 模式: {img.mode}")

# 转换为 RGBA 模式(添加透明通道)
if img.mode != 'RGBA':
    img = img.convert('RGBA')

# 转换为 numpy 数组便于处理
data = np.array(img)
height, width = data.shape[:2]

# 移除右下角水印区域 (假设水印在右下角 1/4 区域)
# 将右下角区域设置为透明或用周围颜色填充
watermark_height = height // 4
watermark_width = width // 4

# 方案1: 直接将右下角设置为透明
data[height - watermark_height:, width - watermark_width:, 3] = 0

# 方案2: 如果需要更精确的处理,可以基于颜色阈值
# 检测白色或特定颜色的水印并设置为透明
# 这里假设水印是白色或浅色
for y in range(0, height):
    for x in range(0, width):
        r, g, b, a = data[y, x]
        # 如果像素接近白色(RGB 值都很高),设置为透明
        if r > 200 and g > 200 and b > 200:
            data[y, x, 3] = 0
        else:
            break

    for x in range(width - 1, 0, -1):
        r, g, b, a = data[y, x]
        # 如果像素接近白色(RGB 值都很高),设置为透明
        if r > 200 and g > 200 and b > 200:
            data[y, x, 3] = 0
        else:
            break

for x in range(0, width):
    for y in range(0, height):
        r, g, b, a = data[y, x]
        # 如果像素接近白色(RGB 值都很高),设置为透明
        if r > 200 and g > 200 and b > 200:
            data[y, x, 3] = 0
        else:
            break

    for y in range(height - 1, 0, -1):
        r, g, b, a = data[y, x]
        # 如果像素接近白色(RGB 值都很高),设置为透明
        if r > 200 and g > 200 and b > 200:
            data[y, x, 3] = 0
        else:
            break

# 转换回 PIL Image
result = Image.fromarray(data, 'RGBA')

# 保存处理后的图片
result.save('chis_processed.png')
print(f"处理完成: chis_processed.png")
print(f"新图片尺寸: {result.size}, 模式: {result.mode}")
