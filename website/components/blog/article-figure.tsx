import Image from 'next/image';
import type { BlogImage } from '@/lib/blog';

const placementClasses: Record<BlogImage['placement'], string> = {
  full: 'article-figure--full',
  'float-left': 'article-figure--float-left',
  'float-right': 'article-figure--float-right',
};

export function ArticleFigure({
  image,
  className = '',
  preload = false,
}: {
  image: BlogImage;
  className?: string;
  preload?: boolean;
}) {
  const sizes = image.placement === 'full'
    ? '(max-width: 767px) calc(100vw - 2rem), 1040px'
    : '(max-width: 767px) calc(100vw - 2rem), 320px';

  return (
    <figure
      className={`article-figure mb-4 border-y border-line-structure ${placementClasses[image.placement]} ${className}`}
      data-blog-image={image.src}
      data-image-placement={image.placement}
    >
      <Image
        src={image.src}
        alt={image.alt}
        width={image.width}
        height={image.height}
        sizes={sizes}
        preload={preload}
        className="h-auto w-full object-cover"
      />
      <figcaption className="border-t border-line-structure bg-surface-1 px-3 py-2 text-[11px] leading-relaxed text-text-disabled">
        {image.caption}
      </figcaption>
    </figure>
  );
}
