// plugins/jsvectormap.client.ts
export default defineNuxtPlugin(() => {
  // Dynamic import to avoid SSR issues
  const loadJsVectorMap = async () => {
    const { default: jsVectorMap } = await import('jsvectormap')
    // await import('jsvectormap/dist/css/jsvectormap.css')
    await import('jsvectormap/dist/maps/world')
    return jsVectorMap
  }

  return {
    provide: {
      loadJsVectorMap
    }
  }
})