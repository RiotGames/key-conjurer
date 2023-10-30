declare module "*.png" {
  const url: string;
  export = url;
}

declare module "*.md" {
  const content: string;
  export = content;
}
