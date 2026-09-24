const SITE = process.env.NEXT_PUBLIC_SITE_URL || "https://sodienthoai.com";
const SHARDS = 256;

export async function GET() {
  const body = [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
    ...Array.from({length: SHARDS}, (_, id) => `  <sitemap><loc>${SITE}/sitemap/${id}.xml</loc></sitemap>`),
    '</sitemapindex>',
  ].join("\n");

  return new Response(body, {
    headers: {
      "Content-Type": "application/xml; charset=utf-8",
      "Cache-Control": "public, s-maxage=3600, stale-while-revalidate=86400",
    },
  });
}
