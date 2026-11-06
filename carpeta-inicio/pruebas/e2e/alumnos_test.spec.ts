import { expect, test } from '@playwright/test';

import { identificadorUnico, ingresarComoAdministrador, seccionDemostracion } from './ayudas';

test.describe('Historia 1: alta de alumnos', () => {
  test('el administrador registra un alumno y lo encuentra en el listado', async ({ page }) => {
    await ingresarComoAdministrador(page);

    const identificador = identificadorUnico();
    await page.goto('/alumnos/nuevo');
    await page.getByLabel('DNI o código *').fill(identificador);
    await page.getByLabel('Nombres *').fill('Ana Lucía');
    await page.getByLabel('Apellidos *').fill('Ríos Núñez');
    await page.getByLabel('Grado y sección *').selectOption({ label: seccionDemostracion });
    await page.getByRole('button', { name: 'Guardar' }).click();

    await expect(page).toHaveURL(new RegExp(`/alumnos/\\d+\\?ok=creado`));
    await expect(page.getByRole('heading', { name: 'Ríos Núñez, Ana Lucía' })).toBeVisible();
    await expect(page.getByText('El alumno se registró correctamente.')).toBeVisible();

    await page.goto('/alumnos');
    await page.getByLabel('Buscar').fill(identificador);
    await page.getByRole('button', { name: 'Filtrar' }).click();

    await expect(page.getByText('Ríos Núñez, Ana Lucía')).toBeVisible();
    await expect(page.getByText(identificador)).toBeVisible();
  });

  test('rechaza un identificador ya registrado', async ({ page }) => {
    await ingresarComoAdministrador(page);

    const identificador = identificadorUnico();
    for (const intento of [1, 2]) {
      await page.goto('/alumnos/nuevo');
      await page.getByLabel('DNI o código *').fill(identificador);
      await page.getByLabel('Nombres *').fill('Alumno');
      await page.getByLabel('Apellidos *').fill('Repetido ' + intento);
      await page.getByLabel('Grado y sección *').selectOption({ label: seccionDemostracion });
      await page.getByRole('button', { name: 'Guardar' }).click();
    }

    await expect(page.getByText('Ya existe un alumno con este DNI o código.')).toBeVisible();
  });
});

test.describe('Historia 4: importación', () => {
  test('importa filas válidas y reporta cada rechazo', async ({ page }) => {
    await ingresarComoAdministrador(page);

    const valido = identificadorUnico();
    const duplicado = identificadorUnico();
    const csv = [
      'identificador,nombres,apellidos,grado,seccion',
      `${valido},José María,Núñez Ríos,3.°,A`,
      `${duplicado},Luis,Pérez,3.°,A`,
      `${duplicado},Luis,Duplicado,3.°,A`,
      `,Sin,Nombre,3.°,A`,
    ].join('\n');

    await page.goto('/alumnos/importar');
    await page.locator('input[name="archivo"]').setInputFiles({
      name: 'alumnos-e2e.csv',
      mimeType: 'text/csv',
      buffer: Buffer.from(csv, 'utf-8'),
    });
    await page.getByRole('button', { name: 'Importar' }).click();

    await expect(page).toHaveURL(/\/alumnos\/importaciones\/\d+\?ok=importado/);
    await expect(page.getByRole('heading', { name: 'Reporte de importación' })).toBeVisible();
    await expect(page.getByText('alumnos-e2e.csv')).toBeVisible();
    await expect(page.getByRole('cell', { name: 'Identificador repetido en el archivo' })).toHaveCount(2);
    await expect(page.getByRole('cell', { name: 'Campo obligatorio vacío' })).toBeVisible();
    await expect(page.getByText('Alumnos creados')).toBeVisible();

    await page.goto('/alumnos');
    await page.getByLabel('Buscar').fill(valido);
    await page.getByRole('button', { name: 'Filtrar' }).click();
    await expect(page.getByText('Núñez Ríos, José María')).toBeVisible();
  });
});

test.describe('Historia 3: carné imprimible', () => {
  test('genera el carné del alumno con QR vigente', async ({ page }) => {
    await ingresarComoAdministrador(page);

    const identificador = identificadorUnico();
    const apellidos = `Mendoza ${identificador}`;
    await page.goto('/alumnos/nuevo');
    await page.getByLabel('DNI o código *').fill(identificador);
    await page.getByLabel('Nombres *').fill('Carla');
    await page.getByLabel('Apellidos *').fill(apellidos);
    await page.getByLabel('Grado y sección *').selectOption({ label: seccionDemostracion });
    await page.getByRole('button', { name: 'Guardar' }).click();
    await page.getByRole('button', { name: 'Generar QR' }).click();
    await expect(page.getByText('El QR se generó correctamente.')).toBeVisible();

    await page.goto('/carnets');
    await page.getByLabel('Sección').selectOption({ label: seccionDemostracion });
    await page.getByRole('button', { name: 'Generar carnés' }).click();

    await expect(page.getByRole('heading', { name: /Carnés de/ })).toBeVisible();
    const carnet = page.locator('article.carnet', { hasText: `${apellidos}, Carla` });
    await expect(carnet).toBeVisible();
    const imagenQR = carnet.getByRole('img', { name: /Código QR/ });
    await expect(imagenQR).toHaveAttribute('src', /\/qr\/imagen$/);
    await expect(imagenQR).toHaveJSProperty('complete', true);
    expect(await imagenQR.evaluate((imagen) => (imagen as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
  });
});
