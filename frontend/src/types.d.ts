declare module "*.png" {
  const url: string;
  export = url;
}

declare module "*.md" {
  const content: string;
  export = content;
}

declare module '*.module.css' {
  const classes: { [key: string]: string };
  export default classes;
}
